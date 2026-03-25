package reverse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	ReverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	ReverseRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/reverse"
	ReverseObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/reverse"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type WaybillService struct {
	waybillRepo *ReverseRepo.WaybillRepository
	trackRepo   *ReverseRepo.TrackingEventRepository
	resultRepo  *ReverseRepo.WarehouseResultRepository
	emitter     *ReverseObs.Emitter
}

func NewWaybillService(deps *app.Deps) *WaybillService {
	if deps == nil || deps.DB == nil {
		return &WaybillService{}
	}
	return &WaybillService{
		waybillRepo: ReverseRepo.NewWaybillRepository(deps.DB),
		trackRepo:   ReverseRepo.NewTrackingEventRepository(deps.DB),
		resultRepo:  ReverseRepo.NewWarehouseResultRepository(deps.DB),
		emitter:     ReverseObs.NewEmitter(deps.RuntimeLogger(context.Background(), "reverse-waybill", nil)),
	}
}

type CreateWaybillRequest struct {
	OrderID     string         `json:"order_id"`
	AfterSaleID string         `json:"after_sale_id"`
	WaybillNo   string         `json:"waybill_no,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type AppendTrackingRequest struct {
	Status      string         `json:"status"`
	Description string         `json:"description,omitempty"`
	OccurredAt  *time.Time     `json:"occurred_at,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

type RecordWarehouseResultRequest struct {
	Result       string         `json:"result"`
	Disposition  string         `json:"disposition"`
	OperatorID   string         `json:"operator_id,omitempty"`
	Notes        string         `json:"notes,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	WaybillState string         `json:"waybill_state,omitempty"`
}

type WaybillDetail struct {
	Waybill          *ReverseModel.Waybill          `json:"waybill"`
	Tracking         []ReverseModel.TrackingEvent   `json:"tracking"`
	WarehouseResults []ReverseModel.WarehouseResult `json:"warehouse_results"`
	Compensation     map[string]any                 `json:"compensation"`
}

func (s *WaybillService) List(ctx context.Context, tenantUUID string) ([]ReverseModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("reverse service unavailable")
	}
	return s.waybillRepo.List(withTenantContext(ctx, tenantUUID))
}

func (s *WaybillService) Create(ctx context.Context, tenantUUID string, req CreateWaybillRequest) (*ReverseModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("reverse service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.OrderID) == "" || strings.TrimSpace(req.AfterSaleID) == "" {
		return nil, errors.New("order_id/after_sale_id required")
	}
	waybillNo := strings.TrimSpace(req.WaybillNo)
	if waybillNo == "" {
		waybillNo = fmt.Sprintf("RWB%s", time.Now().UTC().Format("20060102150405"))
	}
	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	row := &ReverseModel.Waybill{
		ID:          utils.NewUUID(),
		OrderID:     strings.TrimSpace(req.OrderID),
		AfterSaleID: strings.TrimSpace(req.AfterSaleID),
		WaybillNo:   waybillNo,
		Status:      "created",
		Metadata:    datatypes.JSON(metadata),
	}
	if err := s.waybillRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	s.emit(row.TenantUUID, row.WaybillNo, row.AfterSaleID, row.Status, "reverse_waybill.create", "created")
	return row, nil
}

func (s *WaybillService) Detail(ctx context.Context, tenantUUID, waybillID string) (*WaybillDetail, error) {
	if s == nil || s.waybillRepo == nil || s.trackRepo == nil || s.resultRepo == nil {
		return nil, errors.New("reverse service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	tracking, err := s.trackRepo.ListByWaybillID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	results, err := s.resultRepo.ListByWaybillID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return &WaybillDetail{
		Waybill:          row,
		Tracking:         tracking,
		WarehouseResults: results,
		Compensation:     buildCompensation(row),
	}, nil
}

func (s *WaybillService) AppendTracking(ctx context.Context, tenantUUID, waybillID string, req AppendTrackingRequest) (*ReverseModel.TrackingEvent, error) {
	if s == nil || s.waybillRepo == nil || s.trackRepo == nil {
		return nil, errors.New("reverse service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	next := normalizeReverseStatus(req.Status)
	if !isReverseTransitionAllowed(row.Status, next) {
		return nil, errors.New("invalid reverse waybill status transition")
	}
	payload, _ := jsonBytes(req.Payload, []byte("{}"))
	event := &ReverseModel.TrackingEvent{
		ID:          utils.NewUUID(),
		WaybillID:   row.ID,
		Status:      next,
		Description: strings.TrimSpace(req.Description),
		OccurredAt:  req.OccurredAt,
		Payload:     datatypes.JSON(payload),
	}
	if err := s.trackRepo.Create(ctx, event); err != nil {
		return nil, err
	}
	row.Status = next
	if err := s.waybillRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	s.emit(row.TenantUUID, row.WaybillNo, row.AfterSaleID, row.Status, "reverse_waybill.track", "created")
	return event, nil
}

func (s *WaybillService) RecordWarehouseResult(ctx context.Context, tenantUUID, waybillID string, req RecordWarehouseResultRequest) (*ReverseModel.WarehouseResult, *ReverseModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil || s.resultRepo == nil {
		return nil, nil, errors.New("reverse service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, nil, err
	}
	if row.Status != "received" && row.Status != "closed" {
		return nil, nil, errors.New("warehouse result requires received/closed waybill status")
	}
	result := strings.TrimSpace(strings.ToLower(req.Result))
	disposition := strings.TrimSpace(strings.ToLower(req.Disposition))
	if result == "" || disposition == "" {
		return nil, nil, errors.New("result/disposition required")
	}
	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	record := &ReverseModel.WarehouseResult{
		ID:         utils.NewUUID(),
		WaybillID:  row.ID,
		Result:     result,
		OperatorID: strings.TrimSpace(req.OperatorID),
		Notes:      strings.TrimSpace(req.Notes),
		Metadata:   datatypes.JSON(metadata),
	}
	if err := s.resultRepo.Create(ctx, record); err != nil {
		return nil, nil, err
	}
	row.InspectionResult = result
	row.Disposition = disposition
	row.Status = "closed"
	if err := s.waybillRepo.Save(ctx, row); err != nil {
		return nil, nil, err
	}
	s.emit(row.TenantUUID, row.WaybillNo, row.AfterSaleID, row.Status, "reverse_waybill.warehouse_result", "created")
	return record, row, nil
}

func normalizeReverseStatus(status string) string {
	v := strings.TrimSpace(strings.ToLower(status))
	switch v {
	case "created", "in_transit", "received", "closed":
		return v
	default:
		return "in_transit"
	}
}

func isReverseTransitionAllowed(current, next string) bool {
	current = strings.TrimSpace(strings.ToLower(current))
	next = strings.TrimSpace(strings.ToLower(next))
	if current == "" || next == "" {
		return false
	}
	if current == next {
		return true
	}
	allowed := map[string][]string{
		"created":    {"in_transit"},
		"in_transit": {"received"},
		"received":   {"closed"},
	}
	for _, candidate := range allowed[current] {
		if candidate == next {
			return true
		}
	}
	return false
}

func buildCompensation(row *ReverseModel.Waybill) map[string]any {
	if row == nil {
		return map[string]any{"recommendation": "none"}
	}
	disposition := strings.TrimSpace(strings.ToLower(row.Disposition))
	inspection := strings.TrimSpace(strings.ToLower(row.InspectionResult))
	recommendation := "none"
	if disposition == "compensate" || inspection == "damaged" || inspection == "rejected" {
		recommendation = "refund_or_replacement"
	}
	return map[string]any{
		"recommendation": recommendation,
		"disposition":    row.Disposition,
		"inspection":     row.InspectionResult,
	}
}

func (s *WaybillService) emit(tenantUUID, waybillNo, afterSaleID, status, action, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.Emit(ReverseObs.Event{
		Action:      strings.TrimSpace(action),
		TenantUUID:  strings.TrimSpace(tenantUUID),
		WaybillNo:   strings.TrimSpace(waybillNo),
		AfterSaleID: strings.TrimSpace(afterSaleID),
		Status:      strings.TrimSpace(status),
		Result:      strings.TrimSpace(result),
		EmittedAt:   time.Now().UTC(),
	})
}
