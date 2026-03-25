package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	LogisticsObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics/integrations"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

// WaybillService handles waybill creation/query/tracking flows.
type WaybillService struct {
	waybillRepo *LogisticsRepo.WaybillRepository
	trackRepo   *LogisticsRepo.TrackingEventRepository
	carrierRepo *LogisticsRepo.CarrierRepository
	configSvc   *integrations.ConfigService
	adapters    map[string]integrations.Adapter
	emitter     *LogisticsObs.Emitter
}

func NewWaybillService(deps *app.Deps) *WaybillService {
	if deps == nil || deps.DB == nil {
		return &WaybillService{}
	}
	return &WaybillService{
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		trackRepo:   LogisticsRepo.NewTrackingEventRepository(deps.DB),
		carrierRepo: LogisticsRepo.NewCarrierRepository(deps.DB),
		configSvc:   integrations.NewConfigService(),
		adapters:    integrations.NewAdapterRegistry(),
		emitter:     LogisticsObs.NewEmitter(deps.RuntimeLogger(context.Background(), "logistics-waybill", nil)),
	}
}

type CreateWaybillRequest struct {
	OrderID              string `json:"order_id"`
	CarrierID            string `json:"carrier_id"`
	ServiceCode          string `json:"service_code"`
	WaybillNo            string `json:"waybill_no,omitempty"`
	ManualFallbackReason string `json:"manual_fallback_reason,omitempty"`
}

type AppendTrackingRequest struct {
	EventID     string         `json:"event_id,omitempty"`
	Status      string         `json:"status"`
	Source      string         `json:"source,omitempty"`
	Description string         `json:"description,omitempty"`
	OccurredAt  *time.Time     `json:"occurred_at,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

type WaybillDetail struct {
	Waybill  *LogisticsModel.Waybill        `json:"waybill"`
	Tracking []LogisticsModel.TrackingEvent `json:"tracking"`
}

func (s *WaybillService) List(ctx context.Context, tenantUUID string) ([]LogisticsModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("waybill service unavailable")
	}
	return s.waybillRepo.List(withTenantContext(ctx, tenantUUID))
}

func (s *WaybillService) Create(ctx context.Context, tenantUUID string, req CreateWaybillRequest) (*LogisticsModel.Waybill, string, error) {
	if s == nil || s.waybillRepo == nil || s.carrierRepo == nil {
		return nil, "", errors.New("waybill service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.OrderID) == "" || strings.TrimSpace(req.CarrierID) == "" || strings.TrimSpace(req.ServiceCode) == "" {
		return nil, "", errors.New("order_id/carrier_id/service_code required")
	}
	carrier, err := s.carrierRepo.GetByID(ctx, req.CarrierID)
	if err != nil {
		return nil, "", err
	}
	if !strings.EqualFold(carrier.Status, "active") {
		return nil, "", errors.New("carrier is not active")
	}
	adapter, provider := s.resolveAdapter(carrier)
	resolvedServiceCode := s.configSvc.ResolveServiceCode(carrier, req.ServiceCode)
	adapterResult, adapterErr := adapter.CreateWaybill(ctx, carrier, integrations.WaybillCreateInput{
		OrderID:     strings.TrimSpace(req.OrderID),
		ServiceCode: resolvedServiceCode,
		WaybillNo:   strings.TrimSpace(req.WaybillNo),
		Metadata: map[string]any{
			"provider": provider,
		},
	})
	if adapterErr != nil && strings.TrimSpace(req.ManualFallbackReason) == "" {
		return nil, "", adapterErr
	}
	waybillNo := strings.TrimSpace(req.WaybillNo)
	if adapterResult != nil && strings.TrimSpace(adapterResult.WaybillNo) != "" {
		waybillNo = strings.TrimSpace(adapterResult.WaybillNo)
	}
	if waybillNo == "" {
		waybillNo = fmt.Sprintf("WB%s", time.Now().UTC().Format("20060102150405"))
	}
	metadata := map[string]any{"provider": provider}
	if adapterResult != nil && len(adapterResult.Metadata) > 0 {
		metadata["adapter"] = adapterResult.Metadata
	}
	if adapterErr != nil && strings.TrimSpace(req.ManualFallbackReason) != "" {
		metadata["fallback"] = map[string]any{
			"reason":  strings.TrimSpace(req.ManualFallbackReason),
			"adapter": adapterErr.Error(),
		}
	}
	metadataBytes, _ := jsonBytes(metadata, []byte("{}"))
	wb := &LogisticsModel.Waybill{
		ID:          utils.NewUUID(),
		OrderID:     strings.TrimSpace(req.OrderID),
		CarrierID:   strings.TrimSpace(req.CarrierID),
		ServiceCode: strings.TrimSpace(resolvedServiceCode),
		WaybillNo:   waybillNo,
		Status:      "created",
		Metadata:    datatypes.JSON(metadataBytes),
	}
	if adapterResult != nil && strings.TrimSpace(adapterResult.Status) != "" {
		wb.Status = strings.TrimSpace(adapterResult.Status)
	}
	if err := s.waybillRepo.Create(ctx, wb); err != nil {
		return nil, "", err
	}
	idemStatus := "created"
	s.emitEvent(tenantUUID, wb.WaybillNo, wb.Status, "waybill.create", idemStatus)
	return wb, idemStatus, nil
}

func (s *WaybillService) Detail(ctx context.Context, tenantUUID, waybillID string) (*WaybillDetail, error) {
	if s == nil || s.waybillRepo == nil || s.trackRepo == nil {
		return nil, errors.New("waybill service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	tracks, err := s.trackRepo.ListByWaybillID(ctx, wb.ID)
	if err != nil {
		return nil, err
	}
	return &WaybillDetail{Waybill: wb, Tracking: tracks}, nil
}

func (s *WaybillService) Cancel(ctx context.Context, tenantUUID, waybillID string) (*LogisticsModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("waybill service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	wb.Status = "cancelled"
	if err := s.waybillRepo.Save(ctx, wb); err != nil {
		return nil, err
	}
	s.emitEvent(tenantUUID, wb.WaybillNo, wb.Status, "waybill.cancel", "created")
	return wb, nil
}

func (s *WaybillService) AppendTracking(ctx context.Context, tenantUUID, waybillID string, req AppendTrackingRequest) (*LogisticsModel.TrackingEvent, string, error) {
	if s == nil || s.waybillRepo == nil || s.trackRepo == nil || s.carrierRepo == nil {
		return nil, "", errors.New("waybill service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, "", err
	}
	carrier, err := s.carrierRepo.GetByID(ctx, wb.CarrierID)
	if err != nil {
		return nil, "", err
	}
	adapter, _ := s.resolveAdapter(carrier)
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		eventID = utils.NewUUID()
	}
	if exists, err := s.trackRepo.ExistsByEventKey(ctx, wb.WaybillNo, eventID); err == nil && exists {
		return nil, "replayed", nil
	}
	payload, _ := jsonBytes(req.Payload, []byte("{}"))
	event := &LogisticsModel.TrackingEvent{
		ID:          utils.NewUUID(),
		WaybillID:   wb.ID,
		WaybillNo:   wb.WaybillNo,
		EventID:     eventID,
		Status:      normalizeTrackingStatus(adapter.NormalizeTrackingStatus(req.Status)),
		Source:      defaultString(req.Source, "manual"),
		Description: strings.TrimSpace(req.Description),
		OccurredAt:  req.OccurredAt,
		Payload:     datatypes.JSON(payload),
	}
	if err := s.trackRepo.Create(ctx, event); err != nil {
		if LogisticsRepo.IsDuplicateConstraintError(err) {
			return nil, "replayed", nil
		}
		return nil, "", err
	}
	wb.Status = mergeWaybillStatus(wb.Status, event.Status, event.Source)
	if err := s.waybillRepo.Save(ctx, wb); err != nil {
		return nil, "", err
	}
	s.emitEvent(tenantUUID, wb.WaybillNo, wb.Status, "waybill.track", "created")
	return event, "created", nil
}

func normalizeTrackingStatus(status string) string {
	val := strings.TrimSpace(strings.ToLower(status))
	if val == "" {
		return "in_transit"
	}
	return val
}

func mergeWaybillStatus(current, next, source string) string {
	current = strings.TrimSpace(strings.ToLower(current))
	next = strings.TrimSpace(strings.ToLower(next))
	source = strings.TrimSpace(strings.ToLower(source))
	if next == "" {
		return current
	}
	if current == "delivered" {
		return current
	}
	if source == "manual" && current != "" && current != "created" {
		return current
	}
	return next
}

func (s *WaybillService) emitEvent(tenantUUID, waybillNo, status, action, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.Emit(LogisticsObs.Event{
		Action:     action,
		TenantUUID: strings.TrimSpace(tenantUUID),
		WaybillNo:  strings.TrimSpace(waybillNo),
		Status:     strings.TrimSpace(status),
		Result:     strings.TrimSpace(result),
		EmittedAt:  time.Now().UTC(),
	})
}

func (s *WaybillService) resolveAdapter(carrier *LogisticsModel.Carrier) (integrations.Adapter, string) {
	provider := "self"
	if s != nil && s.configSvc != nil {
		provider = s.configSvc.ResolveProvider(carrier)
	}
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		provider = "self"
	}
	if s != nil && s.adapters != nil {
		if adapter, ok := s.adapters[provider]; ok && adapter != nil {
			return adapter, provider
		}
	}
	return integrations.NewSelfAdapter("self"), "self"
}
