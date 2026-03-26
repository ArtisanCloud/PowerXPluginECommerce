package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BillingCaseService struct {
	caseRepo    *LogisticsRepo.BillingCaseRepository
	waybillRepo *LogisticsRepo.WaybillRepository
}

func NewBillingCaseService(deps *app.Deps) *BillingCaseService {
	if deps == nil || deps.DB == nil {
		return &BillingCaseService{}
	}
	return &BillingCaseService{
		caseRepo:    LogisticsRepo.NewBillingCaseRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
	}
}

type BillingCaseQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

type CreateBillingCaseRequest struct {
	WaybillID string         `json:"waybill_id"`
	Reason    string         `json:"reason,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type TransitionBillingCaseRequest struct {
	Action     string         `json:"action"`
	OperatorID string         `json:"operator_id,omitempty"`
	Note       string         `json:"note,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func (s *BillingCaseService) List(ctx context.Context, tenantUUID string, query BillingCaseQuery) ([]LogisticsModel.BillingCase, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("billing case service unavailable")
	}
	return s.caseRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.BillingCaseFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		Status:    normalizeBillingCaseStatus(query.Status),
	})
}

func (s *BillingCaseService) Create(ctx context.Context, tenantUUID string, req CreateBillingCaseRequest) (*LogisticsModel.BillingCase, string, error) {
	if s == nil || s.caseRepo == nil || s.waybillRepo == nil {
		return nil, "", errors.New("billing case service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.WaybillID) == "" {
		return nil, "", errors.New("waybill_id required")
	}
	waybill, err := s.waybillRepo.GetByID(ctx, req.WaybillID)
	if err != nil {
		return nil, "", err
	}
	if waybill.BillingStatus != "settled" {
		return nil, "", errors.New("only settled waybill can create billing case")
	}
	if absFloat(waybill.FeeDiffAmount) <= 0.0001 {
		return nil, "", errors.New("no fee diff found, billing case not required")
	}
	existed, err := s.caseRepo.GetActiveByWaybillID(ctx, req.WaybillID)
	if err == nil && existed != nil {
		return existed, "replayed", nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	row := &LogisticsModel.BillingCase{
		ID:         utils.NewUUID(),
		WaybillID:  waybill.ID,
		CarrierID:  waybill.CarrierID,
		CaseNo:     buildBillingCaseNo(),
		Status:     "open",
		DiffAmount: waybill.FeeDiffAmount,
		Reason:     strings.TrimSpace(req.Reason),
		Resolution: "",
		Metadata:   datatypes.JSON(metadata),
	}
	if err := s.caseRepo.Create(ctx, row); err != nil {
		return nil, "", err
	}
	return row, "created", nil
}

func (s *BillingCaseService) Transition(ctx context.Context, tenantUUID, caseID string, req TransitionBillingCaseRequest) (*LogisticsModel.BillingCase, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("billing case service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	action := normalizeBillingCaseAction(req.Action)
	if action == "" {
		return nil, errors.New("action must be one of: confirm, appeal, writeoff")
	}
	next, err := nextBillingCaseStatus(row.Status, action)
	if err != nil {
		return nil, err
	}
	row.Status = next
	note := strings.TrimSpace(req.Note)
	if note != "" {
		row.Resolution = note
	}
	if next == "written_off" {
		now := time.Now().UTC()
		row.ClosedAt = &now
	}
	if err := s.caseRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func buildBillingCaseNo() string {
	return fmt.Sprintf("BC%s", time.Now().UTC().Format("20060102150405"))
}

func normalizeBillingCaseStatus(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "open":
		return "open"
	case "confirmed":
		return "confirmed"
	case "appealed":
		return "appealed"
	case "written_off":
		return "written_off"
	default:
		return ""
	}
}

func normalizeBillingCaseAction(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "confirm":
		return "confirm"
	case "appeal":
		return "appeal"
	case "writeoff":
		return "writeoff"
	default:
		return ""
	}
}

func nextBillingCaseStatus(current, action string) (string, error) {
	current = normalizeBillingCaseStatus(current)
	switch action {
	case "confirm":
		if current == "open" || current == "appealed" {
			return "confirmed", nil
		}
	case "appeal":
		if current == "open" || current == "confirmed" {
			return "appealed", nil
		}
	case "writeoff":
		if current == "confirmed" || current == "appealed" {
			return "written_off", nil
		}
	}
	return "", errors.New("invalid billing case transition")
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
