package logistics

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type CapacityForecastService struct {
	forecastRepo *LogisticsRepo.CapacityForecastRepository
	planRepo     *LogisticsRepo.CapacityPlanRepository
	decisionRepo *LogisticsRepo.AllocationDecisionRepository
}

type CapacityForecastQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Status          string `json:"status,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type CapacityForecastGenerateRequest struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	WindowDays      int    `json:"window_days,omitempty"`
	OperatorID      string `json:"operator_id,omitempty"`
}

type CapacityForecastApplyRequest struct {
	ForecastID string `json:"forecast_id"`
	OperatorID string `json:"operator_id,omitempty"`
}

func NewCapacityForecastService(deps *app.Deps) *CapacityForecastService {
	if deps == nil || deps.DB == nil {
		return &CapacityForecastService{}
	}
	return &CapacityForecastService{
		forecastRepo: LogisticsRepo.NewCapacityForecastRepository(deps.DB),
		planRepo:     LogisticsRepo.NewCapacityPlanRepository(deps.DB),
		decisionRepo: LogisticsRepo.NewAllocationDecisionRepository(deps.DB),
	}
}

func (s *CapacityForecastService) List(ctx context.Context, tenantUUID string, query CapacityForecastQuery) ([]LogisticsModel.CapacityForecast, error) {
	if s == nil || s.forecastRepo == nil {
		return nil, errors.New("capacity forecast service unavailable")
	}
	return s.forecastRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.CapacityForecastFilter{
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		Status:          normalizeCapacityForecastStatus(query.Status),
		Limit:           query.Limit,
	})
}

func (s *CapacityForecastService) Generate(ctx context.Context, tenantUUID string, req CapacityForecastGenerateRequest) ([]LogisticsModel.CapacityForecast, error) {
	if s == nil || s.forecastRepo == nil || s.planRepo == nil || s.decisionRepo == nil {
		return nil, errors.New("capacity forecast service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowDays := req.WindowDays
	if windowDays <= 0 {
		windowDays = 7
	}
	if windowDays > 90 {
		windowDays = 90
	}
	plans, err := s.planRepo.List(ctx, LogisticsRepo.CapacityPlanFilter{
		CarrierID:       strings.TrimSpace(req.CarrierID),
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		Status:          "active",
		Limit:           200,
	})
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, errors.New("no active capacity plan found")
	}

	decisionRows, err := s.decisionRepo.List(ctx, LogisticsRepo.AllocationDecisionFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	decisionCountByCarrier := map[string]int{}
	for _, row := range decisionRows {
		key := strings.TrimSpace(row.CarrierID)
		if key == "" {
			continue
		}
		decisionCountByCarrier[key]++
	}

	now := time.Now().UTC()
	out := make([]LogisticsModel.CapacityForecast, 0, len(plans))
	for _, plan := range plans {
		current := maxInt(plan.DailyCapacity, 0)
		baseLoad := maxInt(plan.UsedCapacity+plan.ReservedCapacity, 0)
		decisionCount := decisionCountByCarrier[strings.TrimSpace(plan.CarrierID)]
		trendBoost := int(math.Ceil(float64(decisionCount) / float64(maxInt(windowDays, 1))))
		predicted := maxInt(baseLoad+trendBoost, int(math.Round(float64(current)*0.65)))
		target := int(math.Ceil(float64(predicted) * 1.15))
		if target < baseLoad {
			target = baseLoad
		}
		recommended := target
		if recommended < plan.ReservedCapacity {
			recommended = plan.ReservedCapacity
		}

		utilization := 0.0
		if current > 0 {
			utilization = float64(predicted) / float64(current)
		}
		confidence := clampFloat64(55+math.Min(float64(decisionCount), 40), 35, 95)
		riskLevel := "low"
		strategy := "keep_quota"
		if utilization >= 1.1 {
			riskLevel = "high"
			strategy = "increase_quota"
		} else if utilization >= 0.85 {
			riskLevel = "medium"
			strategy = "rebalance_quota"
		} else if utilization < 0.5 {
			strategy = "decrease_quota"
			target = int(math.Max(float64(baseLoad), float64(current)*0.7))
			recommended = target
		}
		metrics, _ := jsonBytes(map[string]any{
			"decision_count": decisionCount,
			"base_load":      baseLoad,
			"utilization":    round2Capacity(utilization * 100),
			"window_days":    windowDays,
		}, []byte("{}"))

		row := LogisticsModel.CapacityForecast{
			ID:                   utils.NewUUID(),
			PlanID:               plan.ID,
			CarrierID:            plan.CarrierID,
			WarehouseID:          plan.WarehouseID,
			DestinationZone:      plan.DestinationZone,
			WindowDays:           windowDays,
			CurrentDailyCapacity: current,
			PredictedDailyVolume: predicted,
			TargetCapacity:       target,
			RecommendedQuota:     recommended,
			Confidence:           round2Capacity(confidence),
			RiskLevel:            riskLevel,
			Strategy:             strategy,
			Status:               "suggested",
			Metrics:              datatypes.JSON(metrics),
			CreatedBy:            strings.TrimSpace(req.OperatorID),
			UpdatedBy:            strings.TrimSpace(req.OperatorID),
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := s.forecastRepo.Save(ctx, &row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func (s *CapacityForecastService) Apply(ctx context.Context, tenantUUID string, req CapacityForecastApplyRequest) (*LogisticsModel.CapacityForecast, error) {
	if s == nil || s.forecastRepo == nil || s.planRepo == nil {
		return nil, errors.New("capacity forecast service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	forecastID := strings.TrimSpace(req.ForecastID)
	if forecastID == "" {
		return nil, errors.New("forecast_id is required")
	}
	row, err := s.forecastRepo.GetByID(ctx, forecastID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(row.Status, "applied") {
		return row, nil
	}
	plan, err := s.planRepo.GetByID(ctx, row.PlanID)
	if err != nil {
		return nil, err
	}
	target := maxInt(row.TargetCapacity, plan.ReservedCapacity)
	if target <= 0 {
		target = plan.ReservedCapacity
	}
	if target > plan.DailyCapacity {
		plan.DailyCapacity = target
	}
	plan.ReservedCapacity = target
	if plan.UsedCapacity > plan.DailyCapacity {
		plan.UsedCapacity = plan.DailyCapacity
	}
	if err := s.planRepo.Save(ctx, plan); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	row.Status = "applied"
	row.AppliedAt = &now
	row.UpdatedBy = strings.TrimSpace(req.OperatorID)
	if err := s.forecastRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func normalizeCapacityForecastStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "suggested", "applied", "dismissed":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return ""
	}
}

func clampFloat64(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2Capacity(v float64) float64 {
	return math.Round(v*100) / 100
}
