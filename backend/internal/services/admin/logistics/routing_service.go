package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type RoutingService struct {
	ruleRepo    *LogisticsRepo.RoutingRuleRepository
	carrierRepo *LogisticsRepo.CarrierRepository
}

func NewRoutingService(deps *app.Deps) *RoutingService {
	if deps == nil || deps.DB == nil {
		return &RoutingService{}
	}
	return &RoutingService{
		ruleRepo:    LogisticsRepo.NewRoutingRuleRepository(deps.DB),
		carrierRepo: LogisticsRepo.NewCarrierRepository(deps.DB),
	}
}

type UpsertRoutingRuleRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	CarrierID       string         `json:"carrier_id"`
	ServiceCode     string         `json:"service_code,omitempty"`
	Priority        int            `json:"priority,omitempty"`
	MinWeight       float64        `json:"min_weight,omitempty"`
	MaxWeight       float64        `json:"max_weight,omitempty"`
	MinOrderAmount  float64        `json:"min_order_amount,omitempty"`
	MaxOrderAmount  float64        `json:"max_order_amount,omitempty"`
	Fallback        *bool          `json:"fallback,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	RuleConfig      map[string]any `json:"rule_config,omitempty"`
}

type PreviewRoutingRequest struct {
	WarehouseID        string  `json:"warehouse_id,omitempty"`
	DestinationZone    string  `json:"destination_zone,omitempty"`
	Weight             float64 `json:"weight,omitempty"`
	OrderAmount        float64 `json:"order_amount,omitempty"`
	PreferredCarrierID string  `json:"preferred_carrier_id,omitempty"`
	ServiceCode        string  `json:"service_code,omitempty"`
}

type RoutingCandidate struct {
	CarrierID   string  `json:"carrier_id"`
	CarrierName string  `json:"carrier_name"`
	ServiceCode string  `json:"service_code"`
	Priority    int     `json:"priority"`
	Score       float64 `json:"score"`
	MatchedRule string  `json:"matched_rule,omitempty"`
}

type RoutingPreviewResult struct {
	WarehouseID     string             `json:"warehouse_id,omitempty"`
	DestinationZone string             `json:"destination_zone,omitempty"`
	CarrierID       string             `json:"carrier_id"`
	CarrierName     string             `json:"carrier_name"`
	ServiceCode     string             `json:"service_code"`
	MatchedRuleID   string             `json:"matched_rule_id,omitempty"`
	MatchedRuleName string             `json:"matched_rule_name,omitempty"`
	Strategy        string             `json:"strategy"`
	Fallback        bool               `json:"fallback"`
	Reason          string             `json:"reason"`
	Candidates      []RoutingCandidate `json:"candidates,omitempty"`
}

func (s *RoutingService) ListRules(ctx context.Context, tenantUUID string) ([]LogisticsModel.RoutingRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("routing service unavailable")
	}
	return s.ruleRepo.List(withTenantContext(ctx, tenantUUID), false)
}

func (s *RoutingService) UpsertRule(ctx context.Context, tenantUUID string, req UpsertRoutingRuleRequest) (*LogisticsModel.RoutingRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("routing service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	if strings.TrimSpace(req.CarrierID) == "" {
		return nil, errors.New("carrier_id required")
	}
	serviceCode := strings.TrimSpace(req.ServiceCode)
	if serviceCode == "" {
		serviceCode = "std"
	}
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	fallback := false
	if req.Fallback != nil {
		fallback = *req.Fallback
	}
	payload, _ := jsonBytes(req.RuleConfig, []byte("{}"))

	id := strings.TrimSpace(req.ID)
	if id != "" {
		existed, err := s.ruleRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		existed.Name = name
		existed.WarehouseID = strings.TrimSpace(req.WarehouseID)
		existed.DestinationZone = normalizeDestinationZone(req.DestinationZone)
		existed.CarrierID = strings.TrimSpace(req.CarrierID)
		existed.ServiceCode = serviceCode
		existed.Priority = priority
		existed.MinWeight = normalizeFloat(req.MinWeight)
		existed.MaxWeight = normalizeFloat(req.MaxWeight)
		existed.MinOrderAmount = normalizeFloat(req.MinOrderAmount)
		existed.MaxOrderAmount = normalizeFloat(req.MaxOrderAmount)
		existed.Fallback = fallback
		existed.Enabled = enabled
		existed.RuleConfig = datatypes.JSON(payload)
		if err := s.ruleRepo.Save(ctx, existed); err != nil {
			return nil, err
		}
		return existed, nil
	}
	row := &LogisticsModel.RoutingRule{
		ID:              utils.NewUUID(),
		Name:            name,
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: normalizeDestinationZone(req.DestinationZone),
		CarrierID:       strings.TrimSpace(req.CarrierID),
		ServiceCode:     serviceCode,
		Priority:        priority,
		MinWeight:       normalizeFloat(req.MinWeight),
		MaxWeight:       normalizeFloat(req.MaxWeight),
		MinOrderAmount:  normalizeFloat(req.MinOrderAmount),
		MaxOrderAmount:  normalizeFloat(req.MaxOrderAmount),
		Fallback:        fallback,
		Enabled:         enabled,
		RuleConfig:      datatypes.JSON(payload),
	}
	if err := s.ruleRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *RoutingService) DeleteRule(ctx context.Context, tenantUUID, ruleID string) error {
	if s == nil || s.ruleRepo == nil {
		return errors.New("routing service unavailable")
	}
	return s.ruleRepo.DeleteByID(withTenantContext(ctx, tenantUUID), ruleID)
}

func (s *RoutingService) Preview(ctx context.Context, tenantUUID string, req PreviewRoutingRequest) (*RoutingPreviewResult, error) {
	if s == nil || s.ruleRepo == nil || s.carrierRepo == nil {
		return nil, errors.New("routing service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	rules, err := s.ruleRepo.List(ctx, true)
	if err != nil {
		return nil, err
	}
	carriers, err := s.carrierRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	carrierByID := map[string]LogisticsModel.Carrier{}
	activeCarriers := make([]LogisticsModel.Carrier, 0, len(carriers))
	for _, c := range carriers {
		carrierByID[c.ID] = c
		if strings.EqualFold(c.Status, "active") {
			activeCarriers = append(activeCarriers, c)
		}
	}

	matched := make([]LogisticsModel.RoutingRule, 0)
	for _, rule := range rules {
		if !ruleMatches(rule, req) {
			continue
		}
		if carrier, ok := carrierByID[rule.CarrierID]; !ok || !strings.EqualFold(carrier.Status, "active") {
			continue
		}
		matched = append(matched, rule)
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Priority != matched[j].Priority {
			return matched[i].Priority > matched[j].Priority
		}
		if matched[i].Fallback != matched[j].Fallback {
			return !matched[i].Fallback && matched[j].Fallback
		}
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	candidates := make([]RoutingCandidate, 0, len(matched))
	for _, rule := range matched {
		carrier := carrierByID[rule.CarrierID]
		score := float64(rule.Priority) + carrierOnTimeRate(carrier.Config)/10
		candidates = append(candidates, RoutingCandidate{
			CarrierID:   carrier.ID,
			CarrierName: carrier.Name,
			ServiceCode: firstNonEmpty(strings.TrimSpace(req.ServiceCode), rule.ServiceCode),
			Priority:    rule.Priority,
			Score:       math.Round(score*100) / 100,
			MatchedRule: rule.ID,
		})
	}

	if len(matched) > 0 {
		first := matched[0]
		carrier := carrierByID[first.CarrierID]
		return &RoutingPreviewResult{
			WarehouseID:     strings.TrimSpace(req.WarehouseID),
			DestinationZone: normalizeDestinationZone(req.DestinationZone),
			CarrierID:       carrier.ID,
			CarrierName:     carrier.Name,
			ServiceCode:     firstNonEmpty(strings.TrimSpace(req.ServiceCode), first.ServiceCode),
			MatchedRuleID:   first.ID,
			MatchedRuleName: first.Name,
			Strategy:        "rule",
			Fallback:        first.Fallback,
			Reason:          "matched by priority rule",
			Candidates:      candidates,
		}, nil
	}

	if pref := strings.TrimSpace(req.PreferredCarrierID); pref != "" {
		if carrier, ok := carrierByID[pref]; ok && strings.EqualFold(carrier.Status, "active") {
			return &RoutingPreviewResult{
				WarehouseID:     strings.TrimSpace(req.WarehouseID),
				DestinationZone: normalizeDestinationZone(req.DestinationZone),
				CarrierID:       carrier.ID,
				CarrierName:     carrier.Name,
				ServiceCode:     firstNonEmpty(strings.TrimSpace(req.ServiceCode), "std"),
				Strategy:        "preferred",
				Fallback:        true,
				Reason:          "fallback to preferred carrier",
			}, nil
		}
	}

	if len(activeCarriers) == 0 {
		return nil, errors.New("no active carriers available")
	}
	sort.Slice(activeCarriers, func(i, j int) bool {
		left := carrierOnTimeRate(activeCarriers[i].Config)
		right := carrierOnTimeRate(activeCarriers[j].Config)
		if left == right {
			return strings.Compare(activeCarriers[i].Name, activeCarriers[j].Name) < 0
		}
		return left > right
	})
	best := activeCarriers[0]
	return &RoutingPreviewResult{
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: normalizeDestinationZone(req.DestinationZone),
		CarrierID:       best.ID,
		CarrierName:     best.Name,
		ServiceCode:     firstNonEmpty(strings.TrimSpace(req.ServiceCode), "std"),
		Strategy:        "fallback",
		Fallback:        true,
		Reason:          "fallback to best on-time active carrier",
	}, nil
}

func ruleMatches(rule LogisticsModel.RoutingRule, req PreviewRoutingRequest) bool {
	if !rule.Enabled {
		return false
	}
	if rule.WarehouseID != "" && !strings.EqualFold(strings.TrimSpace(rule.WarehouseID), strings.TrimSpace(req.WarehouseID)) {
		return false
	}
	zone := normalizeDestinationZone(req.DestinationZone)
	if z := normalizeDestinationZone(rule.DestinationZone); z != "" && z != "GLOBAL" && z != zone {
		return false
	}
	if rule.MinWeight > 0 && req.Weight > 0 && req.Weight < rule.MinWeight {
		return false
	}
	if rule.MaxWeight > 0 && req.Weight > 0 && req.Weight > rule.MaxWeight {
		return false
	}
	if rule.MinOrderAmount > 0 && req.OrderAmount > 0 && req.OrderAmount < rule.MinOrderAmount {
		return false
	}
	if rule.MaxOrderAmount > 0 && req.OrderAmount > 0 && req.OrderAmount > rule.MaxOrderAmount {
		return false
	}
	if svc := strings.TrimSpace(req.ServiceCode); svc != "" && strings.TrimSpace(rule.ServiceCode) != "" && !strings.EqualFold(rule.ServiceCode, svc) {
		return false
	}
	return true
}

func carrierOnTimeRate(cfg datatypes.JSON) float64 {
	if len(cfg) == 0 {
		return 0
	}
	obj := map[string]any{}
	if err := json.Unmarshal(cfg, &obj); err != nil {
		return 0
	}
	if v, ok := obj["onTimeRate"]; ok {
		return toFloat(v)
	}
	if v, ok := obj["on_time_rate"]; ok {
		return toFloat(v)
	}
	return 0
}

func normalizeDestinationZone(v string) string {
	zone := strings.ToUpper(strings.TrimSpace(v))
	if zone == "" {
		return "GLOBAL"
	}
	return zone
}

func normalizeFloat(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		if strings.TrimSpace(n) == "" {
			return 0
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
