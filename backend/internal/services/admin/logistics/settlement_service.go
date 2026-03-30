package logistics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type SettlementService struct {
	batchRepo   *LogisticsRepo.SettlementBatchRepository
	diffRepo    *LogisticsRepo.SettlementDiffRepository
	billingRepo *LogisticsRepo.BillingRepository
}

type CreateSettlementBatchRequest struct {
	CarrierID string `json:"carrier_id,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
}

type HandleSettlementDiffRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Note       string `json:"note,omitempty"`
}

func NewSettlementService(deps *app.Deps) *SettlementService {
	if deps == nil || deps.DB == nil {
		return &SettlementService{}
	}
	wbRepo := LogisticsRepo.NewWaybillRepository(deps.DB)
	return &SettlementService{
		batchRepo:   LogisticsRepo.NewSettlementBatchRepository(deps.DB),
		diffRepo:    LogisticsRepo.NewSettlementDiffRepository(deps.DB),
		billingRepo: LogisticsRepo.NewBillingRepository(wbRepo),
	}
}

func (s *SettlementService) ListBatches(ctx context.Context, tenantUUID, carrierID, status string, limit int) ([]LogisticsModel.SettlementBatch, error) {
	if s == nil || s.batchRepo == nil {
		return nil, errors.New("settlement service unavailable")
	}
	return s.batchRepo.List(withTenantContext(ctx, tenantUUID), carrierID, normalizeSettlementBatchStatus(status), limit)
}

func (s *SettlementService) ListDiffs(ctx context.Context, tenantUUID, batchID, status string, limit int) ([]LogisticsModel.SettlementDiff, error) {
	if s == nil || s.diffRepo == nil {
		return nil, errors.New("settlement service unavailable")
	}
	return s.diffRepo.List(withTenantContext(ctx, tenantUUID), batchID, normalizeSettlementDiffStatus(status), limit)
}

func (s *SettlementService) CreateBatch(ctx context.Context, tenantUUID string, req CreateSettlementBatchRequest) (*LogisticsModel.SettlementBatch, error) {
	if s == nil || s.batchRepo == nil || s.diffRepo == nil || s.billingRepo == nil {
		return nil, errors.New("settlement service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	filter, err := parseBillingFilter(BillingQuery(req))
	if err != nil {
		return nil, err
	}
	items, err := s.billingRepo.ListWaybills(ctx, filter)
	if err != nil {
		return nil, err
	}
	batch := &LogisticsModel.SettlementBatch{
		ID:                utils.NewUUID(),
		BatchNo:           buildSettlementBatchNo(),
		CarrierID:         strings.TrimSpace(req.CarrierID),
		Status:            "open",
		WaybillCount:      len(items),
		DiffCount:         0,
		TotalExpectedFee:  0,
		TotalActualFee:    0,
		TotalDiffAmount:   0,
		SuggestionSummary: datatypes.JSON([]byte("{}")),
		Metadata:          datatypes.JSON([]byte("{}")),
	}
	diffRows := make([]LogisticsModel.SettlementDiff, 0)
	suggestCounter := map[string]int{}
	for _, item := range items {
		batch.TotalExpectedFee += item.EstimatedFee
		batch.TotalActualFee += item.ActualFee
		batch.TotalDiffAmount += item.DiffFee
		attrib, suggestion := buildSettlementAttribution(item.EstimatedFee, item.ActualFee, item.DiffFee)
		status := "pending"
		if suggestion == "auto_confirm" {
			status = "resolved"
		}
		if status == "pending" {
			batch.DiffCount++
		}
		suggestCounter[suggestion]++
		diffRows = append(diffRows, LogisticsModel.SettlementDiff{
			ID:          utils.NewUUID(),
			BatchID:     batch.ID,
			WaybillID:   item.WaybillID,
			WaybillNo:   item.WaybillNo,
			CarrierID:   item.CarrierID,
			ExpectedFee: item.EstimatedFee,
			ActualFee:   item.ActualFee,
			DiffAmount:  item.DiffFee,
			Attribution: attrib,
			Suggestion:  suggestion,
			Status:      status,
			Metadata:    datatypes.JSON([]byte("{}")),
		})
	}
	summaryJSON, _ := jsonBytes(suggestCounter, []byte("{}"))
	batch.SuggestionSummary = datatypes.JSON(summaryJSON)
	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, err
	}
	if err := s.diffRepo.CreateBatch(ctx, diffRows); err != nil {
		return nil, err
	}
	return batch, nil
}

func (s *SettlementService) HandleDiff(ctx context.Context, tenantUUID, diffID string, req HandleSettlementDiffRequest) (*LogisticsModel.SettlementDiff, error) {
	if s == nil || s.diffRepo == nil {
		return nil, errors.New("settlement service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.diffRepo.GetByID(ctx, diffID)
	if err != nil {
		return nil, err
	}
	action := normalizeSettlementDiffAction(req.Action)
	if action == "" {
		return nil, errors.New("action must be accept or dispute")
	}
	row.HandledAction = action
	row.HandledBy = strings.TrimSpace(req.OperatorID)
	row.HandledNote = strings.TrimSpace(req.Note)
	now := time.Now().UTC()
	row.HandledAt = &now
	row.Status = "resolved"
	if err := s.diffRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *SettlementService) ConfirmBatch(ctx context.Context, tenantUUID, batchID string) (*LogisticsModel.SettlementBatch, error) {
	if s == nil || s.batchRepo == nil || s.diffRepo == nil {
		return nil, errors.New("settlement service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	batch, err := s.batchRepo.GetByID(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(batch.Status, "confirmed") {
		return batch, nil
	}
	diffs, err := s.diffRepo.ListByBatchID(ctx, batch.ID)
	if err != nil {
		return nil, err
	}
	for _, diff := range diffs {
		if diff.Status != "resolved" {
			return nil, errors.New("cannot confirm batch with pending diffs")
		}
	}
	batch.Status = "confirmed"
	now := time.Now().UTC()
	batch.ConfirmedAt = &now
	if err := s.batchRepo.Save(ctx, batch); err != nil {
		return nil, err
	}
	return batch, nil
}

func buildSettlementBatchNo() string {
	return fmt.Sprintf("SB%s", time.Now().UTC().Format("20060102150405"))
}

func buildSettlementAttribution(expected, actual, diff float64) (string, string) {
	if math.Abs(diff) <= 0.0001 {
		return "matched", "auto_confirm"
	}
	if math.Abs(actual) <= 0.0001 {
		return "missing_actual_fee", "fetch_invoice"
	}
	if expected > 0 && diff > 0 && (diff/expected) >= 0.2 {
		return "weight_mismatch", "raise_dispute"
	}
	if diff < 0 {
		return "discount_or_override", "accept"
	}
	return "normal_variance", "accept"
}

func normalizeSettlementBatchStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "open":
		return "open"
	case "confirmed":
		return "confirmed"
	default:
		return ""
	}
}

func normalizeSettlementDiffStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pending":
		return "pending"
	case "resolved":
		return "resolved"
	default:
		return ""
	}
}

func normalizeSettlementDiffAction(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "accept":
		return "accept"
	case "dispute":
		return "dispute"
	default:
		return ""
	}
}
