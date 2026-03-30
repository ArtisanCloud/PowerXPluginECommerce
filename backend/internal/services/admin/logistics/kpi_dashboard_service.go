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
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type KPIDashboardService struct {
	waybillRepo *LogisticsRepo.WaybillRepository
	snapRepo    *LogisticsRepo.KPISnapshotRepository
}

type KPIDashboardQuery struct {
	WindowHours     int    `json:"window_hours,omitempty"`
	Dimension       string `json:"dimension,omitempty"`
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type KPIDashboardOverview struct {
	WindowHours         int     `json:"window_hours"`
	Dimension           string  `json:"dimension"`
	TotalWaybills       int     `json:"total_waybills"`
	DeliveredCount      int     `json:"delivered_count"`
	ExceptionCount      int     `json:"exception_count"`
	TimeoutCount        int     `json:"timeout_count"`
	OnTimeRate          float64 `json:"on_time_rate"`
	DeliverySuccessRate float64 `json:"delivery_success_rate"`
	AvgTransitHours     float64 `json:"avg_transit_hours"`
	TotalCost           float64 `json:"total_cost"`
	AvgCost             float64 `json:"avg_cost"`
}

type KPIDashboardTrendItem struct {
	DimensionKey        string  `json:"dimension_key"`
	TotalWaybills       int     `json:"total_waybills"`
	DeliveredCount      int     `json:"delivered_count"`
	ExceptionCount      int     `json:"exception_count"`
	OnTimeRate          float64 `json:"on_time_rate"`
	DeliverySuccessRate float64 `json:"delivery_success_rate"`
	AvgTransitHours     float64 `json:"avg_transit_hours"`
	TotalCost           float64 `json:"total_cost"`
	AvgCost             float64 `json:"avg_cost"`
}

type KPIDashboardDrilldownItem struct {
	WaybillID        string  `json:"waybill_id"`
	WaybillNo        string  `json:"waybill_no"`
	CarrierID        string  `json:"carrier_id"`
	Status           string  `json:"status"`
	WarehouseID      string  `json:"warehouse_id,omitempty"`
	DestinationZone  string  `json:"destination_zone,omitempty"`
	ElapsedHours     int     `json:"elapsed_hours"`
	CostAmount       float64 `json:"cost_amount"`
	TimeoutRiskLevel string  `json:"timeout_risk_level"`
}

func NewKPIDashboardService(deps *app.Deps) *KPIDashboardService {
	if deps == nil || deps.DB == nil {
		return &KPIDashboardService{}
	}
	return &KPIDashboardService{
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		snapRepo:    LogisticsRepo.NewKPISnapshotRepository(deps.DB),
	}
}

func (s *KPIDashboardService) Overview(ctx context.Context, tenantUUID string, query KPIDashboardQuery) (*KPIDashboardOverview, error) {
	rows, windowHours, err := s.listScopedWaybills(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	overview := summarizeKPI(rows, windowHours, normalizeKPIDimension(query.Dimension))
	s.persistSnapshot(ctx, tenantUUID, overview)
	return overview, nil
}

func (s *KPIDashboardService) Trends(ctx context.Context, tenantUUID string, query KPIDashboardQuery) ([]KPIDashboardTrendItem, error) {
	rows, _, err := s.listScopedWaybills(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	dimension := normalizeKPIDimension(query.Dimension)
	if dimension == "" {
		dimension = "carrier"
	}
	buckets := map[string][]LogisticsModel.Waybill{}
	for _, row := range rows {
		key := resolveKPIDimensionKey(row, dimension)
		buckets[key] = append(buckets[key], row)
	}
	items := make([]KPIDashboardTrendItem, 0, len(buckets))
	for key, group := range buckets {
		summary := summarizeKPI(group, normalizeWindowHours(query.WindowHours), dimension)
		items = append(items, KPIDashboardTrendItem{
			DimensionKey:        key,
			TotalWaybills:       summary.TotalWaybills,
			DeliveredCount:      summary.DeliveredCount,
			ExceptionCount:      summary.ExceptionCount,
			OnTimeRate:          summary.OnTimeRate,
			DeliverySuccessRate: summary.DeliverySuccessRate,
			AvgTransitHours:     summary.AvgTransitHours,
			TotalCost:           summary.TotalCost,
			AvgCost:             summary.AvgCost,
		})
	}
	return items, nil
}

func (s *KPIDashboardService) Drilldown(ctx context.Context, tenantUUID string, query KPIDashboardQuery) ([]KPIDashboardDrilldownItem, error) {
	rows, _, err := s.listScopedWaybills(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	now := time.Now().UTC()
	items := make([]KPIDashboardDrilldownItem, 0, limit)
	for _, row := range rows {
		if len(items) >= limit {
			break
		}
		meta := parseWaybillMetadata(row.Metadata)
		cost := row.ActualFeeAmount
		if cost <= 0 {
			cost = row.FeeAmount
		}
		elapsed := int(now.Sub(row.CreatedAt).Hours())
		risk := "low"
		if elapsed >= 72 {
			risk = "high"
		} else if elapsed >= 48 {
			risk = "medium"
		}
		items = append(items, KPIDashboardDrilldownItem{
			WaybillID:        row.ID,
			WaybillNo:        row.WaybillNo,
			CarrierID:        row.CarrierID,
			Status:           row.Status,
			WarehouseID:      meta.WarehouseID,
			DestinationZone:  meta.DestinationZone,
			ElapsedHours:     elapsed,
			CostAmount:       cost,
			TimeoutRiskLevel: risk,
		})
	}
	return items, nil
}

func (s *KPIDashboardService) ExportCSV(ctx context.Context, tenantUUID string, query KPIDashboardQuery) (string, error) {
	rows, err := s.Trends(ctx, tenantUUID, query)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	writer := csv.NewWriter(&b)
	_ = writer.Write([]string{
		"dimension_key", "total_waybills", "delivered_count", "exception_count",
		"on_time_rate", "delivery_success_rate", "avg_transit_hours", "total_cost", "avg_cost",
	})
	for _, row := range rows {
		_ = writer.Write([]string{
			row.DimensionKey,
			strconv.Itoa(row.TotalWaybills),
			strconv.Itoa(row.DeliveredCount),
			strconv.Itoa(row.ExceptionCount),
			fmt.Sprintf("%.2f", row.OnTimeRate),
			fmt.Sprintf("%.2f", row.DeliverySuccessRate),
			fmt.Sprintf("%.2f", row.AvgTransitHours),
			fmt.Sprintf("%.2f", row.TotalCost),
			fmt.Sprintf("%.2f", row.AvgCost),
		})
	}
	writer.Flush()
	return b.String(), writer.Error()
}

func (s *KPIDashboardService) listScopedWaybills(ctx context.Context, tenantUUID string, query KPIDashboardQuery) ([]LogisticsModel.Waybill, int, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, 0, errors.New("kpi dashboard service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowHours := normalizeWindowHours(query.WindowHours)
	windowStart := time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour)
	rows, err := s.waybillRepo.List(ctx)
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]LogisticsModel.Waybill, 0, len(rows))
	for _, row := range rows {
		if row.CreatedAt.UTC().Before(windowStart) {
			continue
		}
		meta := parseWaybillMetadata(row.Metadata)
		if !matchesControlTowerScope(
			row,
			meta,
			query.CarrierID,
			query.WarehouseID,
			query.DestinationZone,
		) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered, windowHours, nil
}

func (s *KPIDashboardService) persistSnapshot(ctx context.Context, tenantUUID string, overview *KPIDashboardOverview) {
	if s == nil || s.snapRepo == nil || overview == nil {
		return
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row := &LogisticsModel.KPISnapshot{
		ID:                  utils.NewUUID(),
		WindowHours:         overview.WindowHours,
		DimensionType:       overview.Dimension,
		DimensionKey:        "all",
		TotalWaybills:       overview.TotalWaybills,
		DeliveredCount:      overview.DeliveredCount,
		ExceptionCount:      overview.ExceptionCount,
		TimeoutCount:        overview.TimeoutCount,
		OnTimeRate:          overview.OnTimeRate,
		DeliverySuccessRate: overview.DeliverySuccessRate,
		AvgTransitHours:     overview.AvgTransitHours,
		AvgCost:             overview.AvgCost,
		TotalCost:           overview.TotalCost,
		Metadata:            datatypes.JSON([]byte("{}")),
	}
	_ = s.snapRepo.Create(ctx, row)
}

func summarizeKPI(rows []LogisticsModel.Waybill, windowHours int, dimension string) *KPIDashboardOverview {
	now := time.Now().UTC()
	out := &KPIDashboardOverview{
		WindowHours: windowHours,
		Dimension:   dimension,
	}
	var transitHours float64
	for _, row := range rows {
		out.TotalWaybills++
		status := strings.ToLower(strings.TrimSpace(row.Status))
		switch status {
		case "delivered", "signed":
			out.DeliveredCount++
		case "delay", "exception", "cancelled":
			out.ExceptionCount++
		}
		if status != "delivered" && status != "signed" && row.CreatedAt.Add(72*time.Hour).Before(now) {
			out.TimeoutCount++
		}
		if status == "delivered" || status == "signed" {
			if row.UpdatedAt.Sub(row.CreatedAt) <= 72*time.Hour {
				out.OnTimeRate += 1
			}
		}
		cost := row.ActualFeeAmount
		if cost <= 0 {
			cost = row.FeeAmount
		}
		out.TotalCost += cost
		transitHours += row.UpdatedAt.Sub(row.CreatedAt).Hours()
	}
	if out.DeliveredCount > 0 {
		out.OnTimeRate = out.OnTimeRate * 100 / float64(out.DeliveredCount)
		out.DeliverySuccessRate = float64(out.DeliveredCount) * 100 / float64(out.TotalWaybills)
	} else if out.TotalWaybills > 0 {
		out.DeliverySuccessRate = 0
	}
	if out.TotalWaybills > 0 {
		out.AvgCost = out.TotalCost / float64(out.TotalWaybills)
		out.AvgTransitHours = transitHours / float64(out.TotalWaybills)
	}
	return out
}

func normalizeKPIDimension(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "tenant", "carrier", "warehouse", "destination_zone":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "carrier"
	}
}

func resolveKPIDimensionKey(row LogisticsModel.Waybill, dimension string) string {
	meta := parseWaybillMetadata(row.Metadata)
	switch dimension {
	case "tenant":
		return strings.TrimSpace(row.TenantUUID)
	case "warehouse":
		if strings.TrimSpace(meta.WarehouseID) != "" {
			return strings.TrimSpace(meta.WarehouseID)
		}
		return "unknown_warehouse"
	case "destination_zone":
		if strings.TrimSpace(meta.DestinationZone) != "" {
			return strings.TrimSpace(meta.DestinationZone)
		}
		return "unknown_zone"
	default:
		if strings.TrimSpace(row.CarrierID) != "" {
			return strings.TrimSpace(row.CarrierID)
		}
		return "unknown_carrier"
	}
}
