package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AllocationService struct {
	planRepo     *LogisticsRepo.CapacityPlanRepository
	decisionRepo *LogisticsRepo.AllocationDecisionRepository
	carrierRepo  *LogisticsRepo.CarrierRepository
}

type UpsertCapacityPlanRequest struct {
	ID               string         `json:"id,omitempty"`
	Name             string         `json:"name"`
	CarrierID        string         `json:"carrier_id"`
	WarehouseID      string         `json:"warehouse_id,omitempty"`
	DestinationZone  string         `json:"destination_zone,omitempty"`
	DailyCapacity    int            `json:"daily_capacity"`
	ReservedCapacity int            `json:"reserved_capacity,omitempty"`
	UsedCapacity     int            `json:"used_capacity,omitempty"`
	Status           string         `json:"status,omitempty"`
	Config           map[string]any `json:"config,omitempty"`
}

type AllocationQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Status          string `json:"status,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type AllocationRequest struct {
	RequestKey       string `json:"request_key,omitempty"`
	WaybillID        string `json:"waybill_id,omitempty"`
	OrderID          string `json:"order_id,omitempty"`
	CarrierID        string `json:"carrier_id,omitempty"`
	WarehouseID      string `json:"warehouse_id,omitempty"`
	DestinationZone  string `json:"destination_zone,omitempty"`
	Strategy         string `json:"strategy,omitempty"`
	OperatorID       string `json:"operator_id,omitempty"`
	PreferredCarrier string `json:"preferred_carrier,omitempty"`
}

type OverrideAllocationRequest struct {
	RequestKey string `json:"request_key,omitempty"`
	DecisionID string `json:"decision_id,omitempty"`
	CarrierID  string `json:"carrier_id"`
	Reason     string `json:"reason,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type AllocationCandidate struct {
	CarrierID       string  `json:"carrier_id"`
	CarrierName     string  `json:"carrier_name"`
	Available       int     `json:"available"`
	UsageRate       float64 `json:"usage_rate"`
	TimelinessScore float64 `json:"timeliness_score"`
	CostScore       float64 `json:"cost_score"`
	FinalScore      float64 `json:"final_score"`
	Reason          string  `json:"reason"`
}

type AllocationResult struct {
	DecisionID     string                `json:"decision_id"`
	RequestKey     string                `json:"request_key"`
	Strategy       string                `json:"strategy"`
	CarrierID      string                `json:"carrier_id"`
	CarrierName    string                `json:"carrier_name"`
	ManualOverride bool                  `json:"manual_override"`
	Reason         string                `json:"reason"`
	Candidates     []AllocationCandidate `json:"candidates"`
	CreatedAt      time.Time             `json:"created_at"`
}

func NewAllocationService(deps *app.Deps) *AllocationService {
	if deps == nil || deps.DB == nil {
		return &AllocationService{}
	}
	return &AllocationService{
		planRepo:     LogisticsRepo.NewCapacityPlanRepository(deps.DB),
		decisionRepo: LogisticsRepo.NewAllocationDecisionRepository(deps.DB),
		carrierRepo:  LogisticsRepo.NewCarrierRepository(deps.DB),
	}
}

func (s *AllocationService) ListPlans(ctx context.Context, tenantUUID string, query AllocationQuery) ([]LogisticsModel.CapacityPlan, error) {
	if s == nil || s.planRepo == nil {
		return nil, errors.New("allocation service unavailable")
	}
	return s.planRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.CapacityPlanFilter{
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		Status:          normalizePlanStatus(query.Status),
		Limit:           query.Limit,
	})
}

func (s *AllocationService) UpsertPlan(ctx context.Context, tenantUUID string, req UpsertCapacityPlanRequest) (*LogisticsModel.CapacityPlan, error) {
	if s == nil || s.planRepo == nil {
		return nil, errors.New("allocation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(req.CarrierID) == "" {
		return nil, errors.New("carrier_id is required")
	}
	if req.DailyCapacity <= 0 {
		return nil, errors.New("daily_capacity must be positive")
	}
	status := normalizePlanStatus(req.Status)
	if status == "" {
		status = "active"
	}
	cfg, err := jsonBytes(req.Config, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row := &LogisticsModel.CapacityPlan{
		ID:               strings.TrimSpace(req.ID),
		Name:             name,
		CarrierID:        strings.TrimSpace(req.CarrierID),
		WarehouseID:      strings.TrimSpace(req.WarehouseID),
		DestinationZone:  strings.TrimSpace(req.DestinationZone),
		DailyCapacity:    req.DailyCapacity,
		ReservedCapacity: req.ReservedCapacity,
		UsedCapacity:     req.UsedCapacity,
		Status:           status,
		Config:           datatypes.JSON(cfg),
	}
	if row.ID == "" {
		existing, getErr := s.planRepo.GetByName(ctx, name)
		if getErr == nil && existing != nil {
			row = existing
			row.CarrierID = strings.TrimSpace(req.CarrierID)
			row.WarehouseID = strings.TrimSpace(req.WarehouseID)
			row.DestinationZone = strings.TrimSpace(req.DestinationZone)
			row.DailyCapacity = req.DailyCapacity
			row.ReservedCapacity = req.ReservedCapacity
			row.UsedCapacity = req.UsedCapacity
			row.Status = status
			row.Config = datatypes.JSON(cfg)
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		} else {
			row.ID = utils.NewUUID()
		}
	} else {
		existing, getErr := s.planRepo.GetByID(ctx, row.ID)
		if getErr == nil && existing != nil {
			row = existing
			row.Name = name
			row.CarrierID = strings.TrimSpace(req.CarrierID)
			row.WarehouseID = strings.TrimSpace(req.WarehouseID)
			row.DestinationZone = strings.TrimSpace(req.DestinationZone)
			row.DailyCapacity = req.DailyCapacity
			row.ReservedCapacity = req.ReservedCapacity
			row.UsedCapacity = req.UsedCapacity
			row.Status = status
			row.Config = datatypes.JSON(cfg)
		}
	}
	if row.ReservedCapacity < 0 {
		row.ReservedCapacity = 0
	}
	if row.UsedCapacity < 0 {
		row.UsedCapacity = 0
	}
	if row.ReservedCapacity > row.DailyCapacity {
		row.ReservedCapacity = row.DailyCapacity
	}
	if err := s.planRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *AllocationService) Allocate(ctx context.Context, tenantUUID string, req AllocationRequest) (*AllocationResult, error) {
	if s == nil || s.planRepo == nil || s.decisionRepo == nil || s.carrierRepo == nil {
		return nil, errors.New("allocation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "alloc#" + utils.NewUUID()
	}
	existing, err := s.decisionRepo.GetByRequestKey(ctx, requestKey)
	if err == nil && existing != nil {
		return s.buildResultFromDecision(existing), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
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
	carriers, err := s.carrierRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	carrierMap := map[string]LogisticsModel.Carrier{}
	for _, c := range carriers {
		carrierMap[c.ID] = c
	}

	candidates := make([]AllocationCandidate, 0, len(plans))
	for _, plan := range plans {
		carrier, ok := carrierMap[plan.CarrierID]
		if !ok {
			continue
		}
		available := plan.DailyCapacity - plan.ReservedCapacity - plan.UsedCapacity
		if available < 0 {
			available = 0
		}
		timeliness := clampScore(carrierOnTimeRate(carrier.Config) / 100)
		cost := clampScore(parseCarrierMetric(carrier.Config, "cost_score", "costScore", 0.5))
		usageRate := 0.0
		if plan.DailyCapacity > 0 {
			usageRate = float64(plan.ReservedCapacity+plan.UsedCapacity) * 100 / float64(plan.DailyCapacity)
		}
		capacityScore := clampScore(float64(available) / float64(maxInt(plan.DailyCapacity, 1)))
		final := 0.45*timeliness + 0.35*capacityScore + 0.20*(1-cost)
		reason := "match"
		if available <= 0 {
			reason = "capacity_exhausted"
		}
		candidates = append(candidates, AllocationCandidate{
			CarrierID:       carrier.ID,
			CarrierName:     carrier.Name,
			Available:       available,
			UsageRate:       round2Optimizer(usageRate),
			TimelinessScore: round2Optimizer(timeliness),
			CostScore:       round2Optimizer(cost),
			FinalScore:      round2Optimizer(final),
			Reason:          reason,
		})
	}
	if len(candidates) == 0 {
		return nil, errors.New("no carrier candidate available")
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Available == candidates[j].Available {
			if candidates[i].FinalScore == candidates[j].FinalScore {
				return candidates[i].CarrierID < candidates[j].CarrierID
			}
			return candidates[i].FinalScore > candidates[j].FinalScore
		}
		return candidates[i].Available > candidates[j].Available
	})

	chosen := candidates[0]
	if pref := strings.TrimSpace(req.PreferredCarrier); pref != "" {
		for _, item := range candidates {
			if item.CarrierID == pref {
				chosen = item
				break
			}
		}
	}
	if chosen.Available <= 0 {
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].CostScore == candidates[j].CostScore {
				return candidates[i].CarrierID < candidates[j].CarrierID
			}
			return candidates[i].CostScore < candidates[j].CostScore
		})
		chosen = candidates[0]
	}
	reason := "capacity+timeliness optimized"
	if strings.TrimSpace(req.PreferredCarrier) != "" && chosen.CarrierID == strings.TrimSpace(req.PreferredCarrier) {
		reason = "preferred carrier override"
	}
	if chosen.Available <= 0 {
		reason = "fallback to lowest_cost due to capacity exhaustion"
	}

	candJSON, _ := jsonBytes(candidates, []byte("[]"))
	decision := &LogisticsModel.AllocationDecision{
		ID:              utils.NewUUID(),
		RequestKey:      requestKey,
		WaybillID:       strings.TrimSpace(req.WaybillID),
		OrderID:         strings.TrimSpace(req.OrderID),
		CarrierID:       chosen.CarrierID,
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		Strategy:        normalizeAllocationStrategy(req.Strategy),
		Reason:          reason,
		ManualOverride:  false,
		Candidates:      datatypes.JSON(candJSON),
		Metadata:        datatypes.JSON([]byte("{}")),
		CreatedBy:       strings.TrimSpace(req.OperatorID),
	}
	if err := s.decisionRepo.Save(ctx, decision); err != nil {
		return nil, err
	}
	if err := s.consumePlanCapacity(ctx, chosen.CarrierID, strings.TrimSpace(req.WarehouseID), strings.TrimSpace(req.DestinationZone)); err != nil {
		return nil, err
	}
	return &AllocationResult{
		DecisionID:     decision.ID,
		RequestKey:     decision.RequestKey,
		Strategy:       decision.Strategy,
		CarrierID:      chosen.CarrierID,
		CarrierName:    chosen.CarrierName,
		ManualOverride: false,
		Reason:         reason,
		Candidates:     candidates,
		CreatedAt:      decision.CreatedAt,
	}, nil
}

func (s *AllocationService) Override(ctx context.Context, tenantUUID string, req OverrideAllocationRequest) (*AllocationResult, error) {
	if s == nil || s.decisionRepo == nil || s.carrierRepo == nil {
		return nil, errors.New("allocation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	carrierID := strings.TrimSpace(req.CarrierID)
	if carrierID == "" {
		return nil, errors.New("carrier_id is required")
	}
	var decision *LogisticsModel.AllocationDecision
	var err error
	if strings.TrimSpace(req.DecisionID) != "" {
		rows, listErr := s.decisionRepo.List(ctx, LogisticsRepo.AllocationDecisionFilter{Limit: 500})
		if listErr != nil {
			return nil, listErr
		}
		for i := range rows {
			if rows[i].ID == strings.TrimSpace(req.DecisionID) {
				tmp := rows[i]
				decision = &tmp
				break
			}
		}
		if decision == nil {
			return nil, gorm.ErrRecordNotFound
		}
	} else if strings.TrimSpace(req.RequestKey) != "" {
		decision, err = s.decisionRepo.GetByRequestKey(ctx, strings.TrimSpace(req.RequestKey))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("decision_id or request_key is required")
	}

	previous := decision.CarrierID
	decision.CarrierID = carrierID
	decision.ManualOverride = true
	decision.PreviousCarrierID = previous
	decision.Reason = strings.TrimSpace(req.Reason)
	if decision.Reason == "" {
		decision.Reason = "manual override"
	}
	if strings.TrimSpace(req.OperatorID) != "" {
		decision.CreatedBy = strings.TrimSpace(req.OperatorID)
	}
	if err := s.decisionRepo.Save(ctx, decision); err != nil {
		return nil, err
	}
	_ = s.consumePlanCapacity(ctx, carrierID, strings.TrimSpace(decision.WarehouseID), strings.TrimSpace(decision.DestinationZone))
	return s.buildResultFromDecision(decision), nil
}

func (s *AllocationService) consumePlanCapacity(ctx context.Context, carrierID, warehouseID, destinationZone string) error {
	plans, err := s.planRepo.List(ctx, LogisticsRepo.CapacityPlanFilter{
		CarrierID:       carrierID,
		WarehouseID:     warehouseID,
		DestinationZone: destinationZone,
		Status:          "active",
		Limit:           1,
	})
	if err != nil || len(plans) == 0 {
		return err
	}
	plan := plans[0]
	plan.UsedCapacity++
	return s.planRepo.Save(ctx, &plan)
}

func (s *AllocationService) buildResultFromDecision(row *LogisticsModel.AllocationDecision) *AllocationResult {
	if row == nil {
		return nil
	}
	candidates := make([]AllocationCandidate, 0)
	if len(row.Candidates) > 0 {
		_ = json.Unmarshal(row.Candidates, &candidates)
	}
	return &AllocationResult{
		DecisionID:     row.ID,
		RequestKey:     row.RequestKey,
		Strategy:       row.Strategy,
		CarrierID:      row.CarrierID,
		CarrierName:    row.CarrierID,
		ManualOverride: row.ManualOverride,
		Reason:         row.Reason,
		Candidates:     candidates,
		CreatedAt:      row.CreatedAt,
	}
}

func normalizePlanStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "active":
		return "active"
	case "paused":
		return "paused"
	default:
		return ""
	}
}

func normalizeAllocationStrategy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "capacity_first":
		return "capacity_first"
	case "timeliness_first":
		return "timeliness_first"
	case "cost_first":
		return "cost_first"
	default:
		return "capacity_first"
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
