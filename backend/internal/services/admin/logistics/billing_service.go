package logistics

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type BillingService struct {
	waybillRepo *LogisticsRepo.WaybillRepository
	billingRepo *LogisticsRepo.BillingRepository
}

func NewBillingService(deps *app.Deps) *BillingService {
	if deps == nil || deps.DB == nil {
		return &BillingService{}
	}
	waybillRepo := LogisticsRepo.NewWaybillRepository(deps.DB)
	return &BillingService{
		waybillRepo: waybillRepo,
		billingRepo: LogisticsRepo.NewBillingRepository(waybillRepo),
	}
}

type UpdateWaybillCostRequest struct {
	ActualFeeAmount float64 `json:"actual_fee_amount"`
}

type BillingQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
}

type BillingSnapshot struct {
	Summary []LogisticsRepo.BillingCarrierSummary `json:"summary"`
	Items   []LogisticsRepo.BillingWaybillItem    `json:"items"`
}

func (s *BillingService) UpdateWaybillCost(ctx context.Context, tenantUUID, waybillID string, req UpdateWaybillCostRequest) (*LogisticsModel.Waybill, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("billing service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	wb.ActualFeeAmount = req.ActualFeeAmount
	wb.FeeDiffAmount = req.ActualFeeAmount - wb.FeeAmount
	wb.BillingStatus = "settled"
	now := time.Now().UTC()
	wb.SettledAt = &now
	if err := s.waybillRepo.Save(ctx, wb); err != nil {
		return nil, err
	}
	return wb, nil
}

func (s *BillingService) Snapshot(ctx context.Context, tenantUUID string, query BillingQuery) (*BillingSnapshot, error) {
	if s == nil || s.billingRepo == nil {
		return nil, errors.New("billing service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	filter, err := parseBillingFilter(query)
	if err != nil {
		return nil, err
	}
	summary, err := s.billingRepo.SummarizeByCarrier(ctx, filter)
	if err != nil {
		return nil, err
	}
	items, err := s.billingRepo.ListWaybills(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &BillingSnapshot{Summary: summary, Items: items}, nil
}

func (s *BillingService) Export(ctx context.Context, tenantUUID string, query BillingQuery, format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		return "", errors.New("format must be csv or json")
	}
	snapshot, err := s.Snapshot(ctx, tenantUUID, query)
	if err != nil {
		return "", err
	}
	if format == "json" {
		return "", nil
	}
	return buildBillingCSV(snapshot.Items), nil
}

func parseBillingFilter(query BillingQuery) (LogisticsRepo.BillingFilter, error) {
	filter := LogisticsRepo.BillingFilter{CarrierID: strings.TrimSpace(query.CarrierID)}
	if strings.TrimSpace(query.From) != "" {
		from, err := time.Parse(time.RFC3339, strings.TrimSpace(query.From))
		if err != nil {
			return filter, errors.New("from must be RFC3339")
		}
		filter.From = &from
	}
	if strings.TrimSpace(query.To) != "" {
		to, err := time.Parse(time.RFC3339, strings.TrimSpace(query.To))
		if err != nil {
			return filter, errors.New("to must be RFC3339")
		}
		filter.To = &to
	}
	return filter, nil
}

func buildBillingCSV(items []LogisticsRepo.BillingWaybillItem) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"waybill_no", "order_id", "carrier", "estimated_fee", "actual_fee", "diff_fee", "billing_status", "created_at"})
	for _, item := range items {
		_ = w.Write([]string{
			item.WaybillNo,
			item.OrderID,
			item.CarrierName,
			strconv.FormatFloat(item.EstimatedFee, 'f', 2, 64),
			strconv.FormatFloat(item.ActualFee, 'f', 2, 64),
			strconv.FormatFloat(item.DiffFee, 'f', 2, 64),
			item.BillingStatus,
			item.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
	return b.String()
}

func (s *BillingService) ExportPayload(ctx context.Context, tenantUUID string, query BillingQuery, format string) (map[string]any, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "csv"
	}
	snapshot, err := s.Snapshot(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	if format == "json" {
		return map[string]any{
			"format":  "json",
			"summary": snapshot.Summary,
			"items":   snapshot.Items,
		}, nil
	}
	if format != "csv" {
		return nil, fmt.Errorf("format must be csv or json")
	}
	return map[string]any{
		"format":  "csv",
		"content": buildBillingCSV(snapshot.Items),
		"count":   len(snapshot.Items),
	}, nil
}
