package logistics

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type InterwarehouseAllocationService struct {
	interRepo *LogisticsRepo.InterwarehouseAllocationRepository
	planRepo  *LogisticsRepo.CapacityPlanRepository
}

type InterwarehouseSuggestionQuery struct {
	RequestKey        string `json:"request_key,omitempty"`
	CarrierID         string `json:"carrier_id,omitempty"`
	SourceWarehouseID string `json:"source_warehouse_id,omitempty"`
	TargetWarehouseID string `json:"target_warehouse_id,omitempty"`
	Status            string `json:"status,omitempty"`
	Limit             int    `json:"limit,omitempty"`
}

type SuggestInterwarehouseAllocationRequest struct {
	RequestKey        string `json:"request_key,omitempty"`
	WaybillID         string `json:"waybill_id,omitempty"`
	OrderID           string `json:"order_id,omitempty"`
	CarrierID         string `json:"carrier_id,omitempty"`
	SourceWarehouseID string `json:"source_warehouse_id,omitempty"`
	DestinationZone   string `json:"destination_zone,omitempty"`
	RequiredQty       int    `json:"required_qty,omitempty"`
	OperatorID        string `json:"operator_id,omitempty"`
}

type ConfirmInterwarehouseAllocationRequest struct {
	CandidateID       string `json:"candidate_id,omitempty"`
	RequestKey        string `json:"request_key,omitempty"`
	TargetWarehouseID string `json:"target_warehouse_id,omitempty"`
	OperatorID        string `json:"operator_id,omitempty"`
}

type ConfirmInterwarehouseAllocationResult struct {
	Selected LogisticsModel.InterwarehouseAllocation   `json:"selected"`
	Items    []LogisticsModel.InterwarehouseAllocation `json:"items"`
}

func NewInterwarehouseAllocationService(deps *app.Deps) *InterwarehouseAllocationService {
	if deps == nil || deps.DB == nil {
		return &InterwarehouseAllocationService{}
	}
	return &InterwarehouseAllocationService{
		interRepo: LogisticsRepo.NewInterwarehouseAllocationRepository(deps.DB),
		planRepo:  LogisticsRepo.NewCapacityPlanRepository(deps.DB),
	}
}

func (s *InterwarehouseAllocationService) List(ctx context.Context, tenantUUID string, query InterwarehouseSuggestionQuery) ([]LogisticsModel.InterwarehouseAllocation, error) {
	if s == nil || s.interRepo == nil {
		return nil, errors.New("interwarehouse allocation service unavailable")
	}
	return s.interRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.InterwarehouseAllocationFilter{
		RequestKey:        strings.TrimSpace(query.RequestKey),
		CarrierID:         strings.TrimSpace(query.CarrierID),
		SourceWarehouseID: strings.TrimSpace(query.SourceWarehouseID),
		TargetWarehouseID: strings.TrimSpace(query.TargetWarehouseID),
		Status:            normalizeInterwarehouseStatus(query.Status),
		Limit:             query.Limit,
	})
}

func (s *InterwarehouseAllocationService) Suggest(ctx context.Context, tenantUUID string, req SuggestInterwarehouseAllocationRequest) ([]LogisticsModel.InterwarehouseAllocation, error) {
	if s == nil || s.interRepo == nil || s.planRepo == nil {
		return nil, errors.New("interwarehouse allocation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "interwh#" + utils.NewUUID()
	}
	if rows, err := s.interRepo.List(ctx, LogisticsRepo.InterwarehouseAllocationFilter{RequestKey: requestKey, Limit: 200}); err == nil && len(rows) > 0 {
		return rows, nil
	}

	requiredQty := req.RequiredQty
	if requiredQty <= 0 {
		requiredQty = 1
	}

	plans, err := s.planRepo.List(ctx, LogisticsRepo.CapacityPlanFilter{
		CarrierID:       strings.TrimSpace(req.CarrierID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		Status:          "active",
		Limit:           500,
	})
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, errors.New("no active capacity plan found")
	}

	sourceWarehouseID := strings.TrimSpace(req.SourceWarehouseID)
	if sourceWarehouseID == "" {
		sourceWarehouseID = strings.TrimSpace(plans[0].WarehouseID)
	}
	var sourcePlan *LogisticsModel.CapacityPlan
	for i := range plans {
		p := plans[i]
		if strings.TrimSpace(p.WarehouseID) == sourceWarehouseID {
			sourcePlan = &p
			break
		}
	}
	if sourcePlan == nil {
		return nil, errors.New("source warehouse capacity plan not found")
	}
	sourceAvailable := calcPlanAvailable(*sourcePlan)
	if sourceAvailable >= requiredQty {
		return nil, errors.New("source warehouse has enough capacity")
	}

	candidates := make([]LogisticsModel.InterwarehouseAllocation, 0)
	now := time.Now().UTC()
	for _, plan := range plans {
		targetWarehouseID := strings.TrimSpace(plan.WarehouseID)
		if targetWarehouseID == "" || targetWarehouseID == sourceWarehouseID {
			continue
		}
		targetAvailable := calcPlanAvailable(plan)
		if targetAvailable <= 0 {
			continue
		}
		transferQty := minInt(requiredQty-sourceAvailable, targetAvailable)
		if transferQty <= 0 {
			continue
		}
		transferCost := estimateTransferCost(sourceWarehouseID, targetWarehouseID, transferQty)
		etaImpact := estimateETAImpact(sourceWarehouseID, targetWarehouseID)
		score := scoreInterwarehouseCandidate(targetAvailable, plan.DailyCapacity, transferCost, etaImpact)
		candidates = append(candidates, LogisticsModel.InterwarehouseAllocation{
			ID:                utils.NewUUID(),
			RequestKey:        requestKey,
			WaybillID:         strings.TrimSpace(req.WaybillID),
			OrderID:           strings.TrimSpace(req.OrderID),
			CarrierID:         strings.TrimSpace(req.CarrierID),
			SourceWarehouseID: sourceWarehouseID,
			TargetWarehouseID: targetWarehouseID,
			DestinationZone:   strings.TrimSpace(req.DestinationZone),
			TransferQty:       transferQty,
			SourceAvailable:   sourceAvailable,
			TargetAvailable:   targetAvailable,
			TransferCost:      round2Interwarehouse(transferCost),
			ETAImpactHours:    round2Interwarehouse(etaImpact),
			Score:             round4Interwarehouse(score),
			Strategy:          "collaboration_first",
			Status:            "suggested",
			Reason:            "source_shortage_detected",
			Metadata:          datatypes.JSON([]byte("{}")),
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}
	if len(candidates) == 0 {
		return nil, errors.New("no interwarehouse candidate available")
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			if candidates[i].TargetAvailable == candidates[j].TargetAvailable {
				return candidates[i].TargetWarehouseID < candidates[j].TargetWarehouseID
			}
			return candidates[i].TargetAvailable > candidates[j].TargetAvailable
		}
		return candidates[i].Score > candidates[j].Score
	})
	if err := s.interRepo.SaveBatch(ctx, candidates); err != nil {
		return nil, err
	}
	return candidates, nil
}

func (s *InterwarehouseAllocationService) Confirm(ctx context.Context, tenantUUID string, req ConfirmInterwarehouseAllocationRequest) (*ConfirmInterwarehouseAllocationResult, error) {
	if s == nil || s.interRepo == nil || s.planRepo == nil {
		return nil, errors.New("interwarehouse allocation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	var selected *LogisticsModel.InterwarehouseAllocation
	var err error
	if strings.TrimSpace(req.CandidateID) != "" {
		selected, err = s.interRepo.GetByID(ctx, strings.TrimSpace(req.CandidateID))
		if err != nil {
			return nil, err
		}
	} else {
		requestKey := strings.TrimSpace(req.RequestKey)
		if requestKey == "" {
			return nil, errors.New("candidate_id or request_key is required")
		}
		rows, listErr := s.interRepo.List(ctx, LogisticsRepo.InterwarehouseAllocationFilter{RequestKey: requestKey, Limit: 200})
		if listErr != nil {
			return nil, listErr
		}
		if len(rows) == 0 {
			return nil, errors.New("no candidate found")
		}
		targetWarehouseID := strings.TrimSpace(req.TargetWarehouseID)
		if targetWarehouseID == "" {
			targetWarehouseID = rows[0].TargetWarehouseID
		}
		for i := range rows {
			if rows[i].TargetWarehouseID == targetWarehouseID {
				tmp := rows[i]
				selected = &tmp
				break
			}
		}
		if selected == nil {
			return nil, errors.New("target warehouse candidate not found")
		}
	}

	rows, err := s.interRepo.List(ctx, LogisticsRepo.InterwarehouseAllocationFilter{RequestKey: selected.RequestKey, Limit: 200})
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(selected.Status, "confirmed") {
		return &ConfirmInterwarehouseAllocationResult{Selected: *selected, Items: rows}, nil
	}

	now := time.Now().UTC()
	for i := range rows {
		if rows[i].ID == selected.ID {
			rows[i].Status = "confirmed"
			rows[i].ConfirmedAt = &now
			rows[i].ConfirmedBy = strings.TrimSpace(req.OperatorID)
			selected = &rows[i]
		} else if rows[i].Status == "suggested" {
			rows[i].Status = "dismissed"
		}
		if saveErr := s.interRepo.Save(ctx, &rows[i]); saveErr != nil {
			return nil, saveErr
		}
	}

	if err := s.applyInterwarehouseCapacity(ctx, *selected); err != nil {
		return nil, err
	}
	return &ConfirmInterwarehouseAllocationResult{Selected: *selected, Items: rows}, nil
}

func (s *InterwarehouseAllocationService) applyInterwarehouseCapacity(ctx context.Context, selected LogisticsModel.InterwarehouseAllocation) error {
	plans, err := s.planRepo.List(ctx, LogisticsRepo.CapacityPlanFilter{
		CarrierID:       strings.TrimSpace(selected.CarrierID),
		DestinationZone: strings.TrimSpace(selected.DestinationZone),
		Status:          "active",
		Limit:           500,
	})
	if err != nil {
		return err
	}
	for i := range plans {
		plan := plans[i]
		switch strings.TrimSpace(plan.WarehouseID) {
		case strings.TrimSpace(selected.SourceWarehouseID):
			plan.UsedCapacity = maxInt(0, plan.UsedCapacity-selected.TransferQty)
			if err := s.planRepo.Save(ctx, &plan); err != nil {
				return err
			}
		case strings.TrimSpace(selected.TargetWarehouseID):
			plan.UsedCapacity += selected.TransferQty
			if plan.UsedCapacity > plan.DailyCapacity {
				plan.UsedCapacity = plan.DailyCapacity
			}
			if err := s.planRepo.Save(ctx, &plan); err != nil {
				return err
			}
		}
	}
	return nil
}

func calcPlanAvailable(plan LogisticsModel.CapacityPlan) int {
	available := plan.DailyCapacity - plan.ReservedCapacity - plan.UsedCapacity
	if available < 0 {
		return 0
	}
	return available
}

func scoreInterwarehouseCandidate(targetAvailable, dailyCapacity int, transferCost, etaImpact float64) float64 {
	availabilityRatio := 0.0
	if dailyCapacity > 0 {
		availabilityRatio = float64(targetAvailable) / float64(dailyCapacity)
	}
	availabilityRatio = clampFloat64(availabilityRatio, 0, 1)
	costScore := 1 - clampFloat64(transferCost/120, 0, 1)
	etaScore := 1 - clampFloat64(etaImpact/36, 0, 1)
	return 0.5*availabilityRatio + 0.3*costScore + 0.2*etaScore
}

func estimateTransferCost(sourceWarehouseID, targetWarehouseID string, qty int) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(sourceWarehouseID + "#" + targetWarehouseID))
	seed := float64(h.Sum32()%20 + 10)
	return seed + float64(maxInt(1, qty))*0.8
}

func estimateETAImpact(sourceWarehouseID, targetWarehouseID string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(targetWarehouseID + "->" + sourceWarehouseID))
	return float64(h.Sum32()%18 + 4)
}

func normalizeInterwarehouseStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "suggested", "confirmed", "dismissed", "cancelled":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func round2Interwarehouse(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4Interwarehouse(v float64) float64 {
	return math.Round(v*10000) / 10000
}
