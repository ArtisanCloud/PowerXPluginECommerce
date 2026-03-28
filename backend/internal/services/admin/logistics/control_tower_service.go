package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type ControlTowerService struct {
	waybillRepo  *LogisticsRepo.WaybillRepository
	snapshotRepo *LogisticsRepo.ControlTowerSnapshotRepository
	subRepo      *LogisticsRepo.ControlTowerAlertSubscriptionRepository
}

type ControlTowerOverviewQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
}

type ControlTowerDrilldownQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Status          string `json:"status,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type UpsertControlTowerSubscriptionRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	CarrierID       string         `json:"carrier_id,omitempty"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	MinOnTimeRate   float64        `json:"min_on_time_rate,omitempty"`
	MaxTimeoutCount int            `json:"max_timeout_count,omitempty"`
	MaxCostAmount   float64        `json:"max_cost_amount,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	Config          map[string]any `json:"config,omitempty"`
}

type ControlTowerSummary struct {
	WindowHours    int     `json:"window_hours"`
	TotalWaybills  int     `json:"total_waybills"`
	InTransitCount int     `json:"in_transit_count"`
	ExceptionCount int     `json:"exception_count"`
	TimeoutCount   int     `json:"timeout_count"`
	DeliveredCount int     `json:"delivered_count"`
	OnTimeRate     float64 `json:"on_time_rate"`
	TotalCost      float64 `json:"total_cost"`
	AlertCount     int     `json:"alert_count"`
}

type ControlTowerAlert struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ControlTowerOverview struct {
	Summary ControlTowerSummary `json:"summary"`
	Alerts  []ControlTowerAlert `json:"alerts"`
}

type ControlTowerDrilldownItem struct {
	WaybillID        string  `json:"waybill_id"`
	WaybillNo        string  `json:"waybill_no"`
	CarrierID        string  `json:"carrier_id"`
	Status           string  `json:"status"`
	WarehouseID      string  `json:"warehouse_id,omitempty"`
	DestinationZone  string  `json:"destination_zone,omitempty"`
	CostAmount       float64 `json:"cost_amount"`
	ElapsedHours     int     `json:"elapsed_hours"`
	TimeoutRiskLevel string  `json:"timeout_risk_level"`
}

func NewControlTowerService(deps *app.Deps) *ControlTowerService {
	if deps == nil || deps.DB == nil {
		return &ControlTowerService{}
	}
	return &ControlTowerService{
		waybillRepo:  LogisticsRepo.NewWaybillRepository(deps.DB),
		snapshotRepo: LogisticsRepo.NewControlTowerSnapshotRepository(deps.DB),
		subRepo:      LogisticsRepo.NewControlTowerAlertSubscriptionRepository(deps.DB),
	}
}

func (s *ControlTowerService) Overview(ctx context.Context, tenantUUID string, query ControlTowerOverviewQuery) (*ControlTowerOverview, error) {
	if s == nil || s.waybillRepo == nil || s.snapshotRepo == nil {
		return nil, errors.New("control tower service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowHours := normalizeWindowHours(query.WindowHours)
	now := time.Now().UTC()
	windowStart := now.Add(-time.Duration(windowHours) * time.Hour)

	rows, err := s.waybillRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	summary := ControlTowerSummary{WindowHours: windowHours}
	for _, row := range rows {
		if row.CreatedAt.UTC().Before(windowStart) {
			continue
		}
		meta := parseWaybillMetadata(row.Metadata)
		if !matchesControlTowerScope(row, meta, query.CarrierID, query.WarehouseID, query.DestinationZone) {
			continue
		}
		summary.TotalWaybills++
		status := strings.ToLower(strings.TrimSpace(row.Status))
		switch status {
		case "delay", "exception", "cancelled":
			summary.ExceptionCount++
		case "delivered", "signed":
			summary.DeliveredCount++
		default:
			summary.InTransitCount++
		}
		if status != "delivered" && status != "signed" {
			if row.CreatedAt.Add(72 * time.Hour).Before(now) {
				summary.TimeoutCount++
			}
		}
		if row.UpdatedAt.Sub(row.CreatedAt) <= 72*time.Hour && (status == "delivered" || status == "signed") {
			summary.OnTimeRate += 1
		}
		if row.ActualFeeAmount > 0 {
			summary.TotalCost += row.ActualFeeAmount
		} else {
			summary.TotalCost += row.FeeAmount
		}
	}
	if summary.DeliveredCount > 0 {
		summary.OnTimeRate = summary.OnTimeRate * 100 / float64(summary.DeliveredCount)
	} else {
		summary.OnTimeRate = 0
	}
	alerts := buildControlTowerAlerts(summary)
	summary.AlertCount = len(alerts)

	snapshot := &LogisticsModel.ControlTowerSnapshot{
		ID:              utils.NewUUID(),
		WindowHours:     windowHours,
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		InTransitCount:  summary.InTransitCount,
		ExceptionCount:  summary.ExceptionCount,
		TimeoutCount:    summary.TimeoutCount,
		DeliveredCount:  summary.DeliveredCount,
		OnTimeRate:      summary.OnTimeRate,
		TotalCost:       summary.TotalCost,
		AlertCount:      summary.AlertCount,
		Summary:         datatypes.JSON([]byte("{}")),
	}
	if data, jsonErr := jsonBytes(summary, []byte("{}")); jsonErr == nil {
		snapshot.Summary = datatypes.JSON(data)
	}
	_ = s.snapshotRepo.Create(ctx, snapshot)

	return &ControlTowerOverview{
		Summary: summary,
		Alerts:  alerts,
	}, nil
}

func (s *ControlTowerService) Drilldown(ctx context.Context, tenantUUID string, query ControlTowerDrilldownQuery) ([]ControlTowerDrilldownItem, error) {
	if s == nil || s.waybillRepo == nil {
		return nil, errors.New("control tower service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowHours := normalizeWindowHours(query.WindowHours)
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	now := time.Now().UTC()
	windowStart := now.Add(-time.Duration(windowHours) * time.Hour)
	rows, err := s.waybillRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	targetStatus := normalizeControlTowerStatus(query.Status)
	items := make([]ControlTowerDrilldownItem, 0, limit)
	for _, row := range rows {
		if len(items) >= limit {
			break
		}
		if row.CreatedAt.UTC().Before(windowStart) {
			continue
		}
		meta := parseWaybillMetadata(row.Metadata)
		if !matchesControlTowerScope(row, meta, query.CarrierID, query.WarehouseID, query.DestinationZone) {
			continue
		}
		if !matchesControlTowerStatus(row.Status, targetStatus, row.CreatedAt, now) {
			continue
		}
		costAmount := row.ActualFeeAmount
		if costAmount <= 0 {
			costAmount = row.FeeAmount
		}
		elapsed := int(now.Sub(row.CreatedAt).Hours())
		riskLevel := "low"
		if elapsed >= 72 {
			riskLevel = "high"
		} else if elapsed >= 48 {
			riskLevel = "medium"
		}
		items = append(items, ControlTowerDrilldownItem{
			WaybillID:        row.ID,
			WaybillNo:        row.WaybillNo,
			CarrierID:        row.CarrierID,
			Status:           row.Status,
			WarehouseID:      meta.WarehouseID,
			DestinationZone:  meta.DestinationZone,
			CostAmount:       costAmount,
			ElapsedHours:     elapsed,
			TimeoutRiskLevel: riskLevel,
		})
	}
	return items, nil
}

func (s *ControlTowerService) ListSubscriptions(ctx context.Context, tenantUUID string, enabled *bool) ([]LogisticsModel.ControlTowerAlertSubscription, error) {
	if s == nil || s.subRepo == nil {
		return nil, errors.New("control tower service unavailable")
	}
	return s.subRepo.List(withTenantContext(ctx, tenantUUID), enabled)
}

func (s *ControlTowerService) UpsertSubscription(
	ctx context.Context,
	tenantUUID string,
	req UpsertControlTowerSubscriptionRequest,
) (*LogisticsModel.ControlTowerAlertSubscription, error) {
	if s == nil || s.subRepo == nil {
		return nil, errors.New("control tower service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	row := &LogisticsModel.ControlTowerAlertSubscription{
		ID:              strings.TrimSpace(req.ID),
		Name:            name,
		CarrierID:       strings.TrimSpace(req.CarrierID),
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		MinOnTimeRate:   req.MinOnTimeRate,
		MaxTimeoutCount: req.MaxTimeoutCount,
		MaxCostAmount:   req.MaxCostAmount,
		Enabled:         true,
		Config:          datatypes.JSON([]byte("{}")),
	}
	if row.ID == "" {
		row.ID = utils.NewUUID()
	} else {
		existing, err := s.subRepo.GetByID(ctx, row.ID)
		if err == nil {
			row = existing
			row.Name = name
			row.CarrierID = strings.TrimSpace(req.CarrierID)
			row.WarehouseID = strings.TrimSpace(req.WarehouseID)
			row.DestinationZone = strings.TrimSpace(req.DestinationZone)
			if req.MinOnTimeRate > 0 {
				row.MinOnTimeRate = req.MinOnTimeRate
			}
			if req.MaxTimeoutCount >= 0 {
				row.MaxTimeoutCount = req.MaxTimeoutCount
			}
			if req.MaxCostAmount >= 0 {
				row.MaxCostAmount = req.MaxCostAmount
			}
		}
	}
	if row.MinOnTimeRate <= 0 {
		row.MinOnTimeRate = 95
	}
	if row.MaxTimeoutCount < 0 {
		row.MaxTimeoutCount = 0
	}
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	}
	configJSON, err := jsonBytes(req.Config, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row.Config = datatypes.JSON(configJSON)
	if err := s.subRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

type waybillMetadata struct {
	WarehouseID     string `json:"warehouse_id"`
	DestinationZone string `json:"destination_zone"`
}

func parseWaybillMetadata(raw datatypes.JSON) waybillMetadata {
	meta := waybillMetadata{}
	if len(raw) == 0 {
		return meta
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return meta
	}
	if v, ok := payload["warehouse_id"].(string); ok {
		meta.WarehouseID = strings.TrimSpace(v)
	}
	if v, ok := payload["destination_zone"].(string); ok {
		meta.DestinationZone = strings.TrimSpace(v)
	}
	return meta
}

func matchesControlTowerScope(
	row LogisticsModel.Waybill,
	meta waybillMetadata,
	carrierID, warehouseID, destinationZone string,
) bool {
	if strings.TrimSpace(carrierID) != "" && row.CarrierID != strings.TrimSpace(carrierID) {
		return false
	}
	if strings.TrimSpace(warehouseID) != "" && meta.WarehouseID != strings.TrimSpace(warehouseID) {
		return false
	}
	if strings.TrimSpace(destinationZone) != "" && meta.DestinationZone != strings.TrimSpace(destinationZone) {
		return false
	}
	return true
}

func normalizeWindowHours(windowHours int) int {
	if windowHours <= 0 {
		return 24
	}
	if windowHours > 24*30 {
		return 24 * 30
	}
	return windowHours
}

func normalizeControlTowerStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "in_transit":
		return "in_transit"
	case "exception":
		return "exception"
	case "timeout":
		return "timeout"
	case "delivered":
		return "delivered"
	default:
		return ""
	}
}

func matchesControlTowerStatus(rawStatus, target string, createdAt, now time.Time) bool {
	status := strings.ToLower(strings.TrimSpace(rawStatus))
	switch target {
	case "in_transit":
		return status != "delivered" && status != "signed" && status != "delay" && status != "exception" && status != "cancelled"
	case "exception":
		return status == "delay" || status == "exception" || status == "cancelled"
	case "timeout":
		return status != "delivered" && status != "signed" && createdAt.Add(72*time.Hour).Before(now)
	case "delivered":
		return status == "delivered" || status == "signed"
	default:
		return true
	}
}

func buildControlTowerAlerts(summary ControlTowerSummary) []ControlTowerAlert {
	alerts := make([]ControlTowerAlert, 0, 4)
	if summary.OnTimeRate > 0 && summary.OnTimeRate < 95 {
		alerts = append(alerts, ControlTowerAlert{
			Level:   "warning",
			Code:    "on_time_rate_low",
			Message: "准时率低于阈值 95%",
		})
	}
	if summary.TimeoutCount >= 10 {
		alerts = append(alerts, ControlTowerAlert{
			Level:   "warning",
			Code:    "timeout_high",
			Message: "超时运单数量持续升高",
		})
	}
	if summary.ExceptionCount >= 10 {
		alerts = append(alerts, ControlTowerAlert{
			Level:   "warning",
			Code:    "exception_high",
			Message: "异常运单数量持续升高",
		})
	}
	if summary.TotalCost >= 5000 {
		alerts = append(alerts, ControlTowerAlert{
			Level:   "info",
			Code:    "cost_high",
			Message: "窗口内履约成本较高，请关注成本结构",
		})
	}
	return alerts
}
