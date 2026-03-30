package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

type RoutingOptimizerService struct {
	profileRepo *LogisticsRepo.RoutingScoreProfileRepository
	simRepo     *LogisticsRepo.RoutingScoreSimulationRepository
	carrierRepo *LogisticsRepo.CarrierRepository
}

type UpsertRoutingOptimizerStrategyRequest struct {
	Name             string         `json:"name"`
	TimelinessWeight float64        `json:"timeliness_weight"`
	CostWeight       float64        `json:"cost_weight"`
	QuotaWeight      float64        `json:"quota_weight"`
	RiskWeight       float64        `json:"risk_weight"`
	FallbackStrategy string         `json:"fallback_strategy,omitempty"`
	Enabled          *bool          `json:"enabled,omitempty"`
	Config           map[string]any `json:"config,omitempty"`
}

type RoutingOptimizerSimulationRequest struct {
	RequestKey          string   `json:"request_key,omitempty"`
	WarehouseID         string   `json:"warehouse_id,omitempty"`
	DestinationZone     string   `json:"destination_zone,omitempty"`
	Weight              float64  `json:"weight,omitempty"`
	OrderAmount         float64  `json:"order_amount,omitempty"`
	PreferredCarrierID  string   `json:"preferred_carrier_id,omitempty"`
	AvailableCarrierIDs []string `json:"available_carrier_ids,omitempty"`
	RealtimeDegraded    bool     `json:"realtime_degraded,omitempty"`
}

type RoutingOptimizerCandidate struct {
	CarrierID   string  `json:"carrier_id"`
	CarrierName string  `json:"carrier_name"`
	FinalScore  float64 `json:"final_score"`
	Timeliness  float64 `json:"timeliness"`
	Cost        float64 `json:"cost"`
	Quota       float64 `json:"quota"`
	Risk        float64 `json:"risk"`
	Explain     string  `json:"explain"`
}

type RoutingOptimizerSimulationResult struct {
	RequestKey  string                      `json:"request_key"`
	ProfileID   string                      `json:"profile_id,omitempty"`
	Strategy    string                      `json:"strategy"`
	Degraded    bool                        `json:"degraded"`
	Reason      string                      `json:"reason"`
	CarrierID   string                      `json:"carrier_id,omitempty"`
	CarrierName string                      `json:"carrier_name,omitempty"`
	Explain     string                      `json:"explain,omitempty"`
	Candidates  []RoutingOptimizerCandidate `json:"candidates,omitempty"`
	CreatedAt   time.Time                   `json:"created_at"`
}

func NewRoutingOptimizerService(deps *app.Deps) *RoutingOptimizerService {
	if deps == nil || deps.DB == nil {
		return &RoutingOptimizerService{}
	}
	return &RoutingOptimizerService{
		profileRepo: LogisticsRepo.NewRoutingScoreProfileRepository(deps.DB),
		simRepo:     LogisticsRepo.NewRoutingScoreSimulationRepository(deps.DB),
		carrierRepo: LogisticsRepo.NewCarrierRepository(deps.DB),
	}
}

func (s *RoutingOptimizerService) GetStrategy(ctx context.Context, tenantUUID string) (*LogisticsModel.RoutingScoreProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("routing optimizer service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.profileRepo.FirstEnabled(ctx)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.ensureDefaultProfile(ctx)
}

func (s *RoutingOptimizerService) UpsertStrategy(ctx context.Context, tenantUUID string, req UpsertRoutingOptimizerStrategyRequest) (*LogisticsModel.RoutingScoreProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("routing optimizer service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "default"
	}
	weights := sanitizeRoutingWeights(req.TimelinessWeight, req.CostWeight, req.QuotaWeight, req.RiskWeight)
	fallback := normalizeFallbackStrategy(req.FallbackStrategy)
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	cfg, _ := jsonBytes(req.Config, []byte("{}"))

	row, err := s.profileRepo.GetByName(ctx, name)
	if err == nil && row != nil {
		row.TimelinessWeight = weights[0]
		row.CostWeight = weights[1]
		row.QuotaWeight = weights[2]
		row.RiskWeight = weights[3]
		row.FallbackStrategy = fallback
		row.Enabled = enabled
		row.Config = datatypes.JSON(cfg)
		if err := s.profileRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := &LogisticsModel.RoutingScoreProfile{
		ID:               utils.NewUUID(),
		Name:             name,
		TimelinessWeight: weights[0],
		CostWeight:       weights[1],
		QuotaWeight:      weights[2],
		RiskWeight:       weights[3],
		FallbackStrategy: fallback,
		Enabled:          enabled,
		Config:           datatypes.JSON(cfg),
	}
	if err := s.profileRepo.Save(ctx, created); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *RoutingOptimizerService) Simulate(ctx context.Context, tenantUUID string, req RoutingOptimizerSimulationRequest) (*RoutingOptimizerSimulationResult, error) {
	if s == nil || s.profileRepo == nil || s.carrierRepo == nil || s.simRepo == nil {
		return nil, errors.New("routing optimizer service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	profile, err := s.GetStrategy(ctx, tenantUUID)
	if err != nil {
		return nil, err
	}
	carriers, err := s.carrierRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	filtered := filterOptimizerCarriers(carriers, req.AvailableCarrierIDs)
	if len(filtered) == 0 {
		return nil, errors.New("no carrier available for simulation")
	}
	weights := sanitizeRoutingWeights(profile.TimelinessWeight, profile.CostWeight, profile.QuotaWeight, profile.RiskWeight)
	candidates := make([]RoutingOptimizerCandidate, 0, len(filtered))
	for _, carrier := range filtered {
		timeliness := clampScore(carrierOnTimeRate(carrier.Config) / 100)
		cost := clampScore(parseCarrierMetric(carrier.Config, "cost_score", "costScore", 0.5))
		quota := clampScore(parseCarrierMetric(carrier.Config, "quota_remain", "quotaRemain", 0.5))
		risk := clampScore(parseCarrierMetric(carrier.Config, "risk_score", "riskScore", 0.5))
		final := weights[0]*timeliness + weights[1]*(1-cost) + weights[2]*quota + weights[3]*(1-risk)
		candidates = append(candidates, RoutingOptimizerCandidate{
			CarrierID:   carrier.ID,
			CarrierName: carrier.Name,
			FinalScore:  round2Optimizer(final),
			Timeliness:  round2Optimizer(timeliness),
			Cost:        round2Optimizer(cost),
			Quota:       round2Optimizer(quota),
			Risk:        round2Optimizer(risk),
			Explain:     buildOptimizerExplain(carrier.Name, timeliness, cost, quota, risk),
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].FinalScore == candidates[j].FinalScore {
			return candidates[i].CarrierID < candidates[j].CarrierID
		}
		return candidates[i].FinalScore > candidates[j].FinalScore
	})

	result := &RoutingOptimizerSimulationResult{
		RequestKey: strings.TrimSpace(req.RequestKey),
		ProfileID:  profile.ID,
		Strategy:   "optimizer",
		Degraded:   req.RealtimeDegraded,
		Reason:     "multi-objective score selected",
		CreatedAt:  time.Now().UTC(),
		Candidates: candidates,
	}
	if result.RequestKey == "" {
		result.RequestKey = "optimizer#" + utils.NewUUID()
	}
	if req.RealtimeDegraded {
		result.Strategy = "degraded"
		result.Reason = "realtime metric unavailable, using fallback strategy"
		s.applyFallback(profile.FallbackStrategy, candidates)
	}
	if len(candidates) > 0 {
		result.CarrierID = candidates[0].CarrierID
		result.CarrierName = candidates[0].CarrierName
		result.Explain = candidates[0].Explain
	}
	if pref := strings.TrimSpace(req.PreferredCarrierID); pref != "" {
		for idx := range candidates {
			if candidates[idx].CarrierID == pref {
				result.Reason = "preferred carrier override"
				result.CarrierID = candidates[idx].CarrierID
				result.CarrierName = candidates[idx].CarrierName
				result.Explain = candidates[idx].Explain
				break
			}
		}
	}
	if err := s.persistSimulation(ctx, profile.ID, result, req); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RoutingOptimizerService) persistSimulation(ctx context.Context, profileID string, result *RoutingOptimizerSimulationResult, req RoutingOptimizerSimulationRequest) error {
	inPayload, _ := jsonBytes(req, []byte("{}"))
	candidatesPayload, _ := jsonBytes(result.Candidates, []byte("[]"))
	explainPayload, _ := jsonBytes(map[string]any{
		"carrier_id":   result.CarrierID,
		"carrier_name": result.CarrierName,
		"reason":       result.Reason,
		"explain":      result.Explain,
	}, []byte("{}"))
	row := &LogisticsModel.RoutingScoreSimulation{
		ID:             utils.NewUUID(),
		RequestKey:     result.RequestKey,
		ProfileID:      profileID,
		PreferredID:    strings.TrimSpace(req.PreferredCarrierID),
		ChosenID:       result.CarrierID,
		ChosenName:     result.CarrierName,
		Strategy:       result.Strategy,
		Degraded:       result.Degraded,
		Reason:         result.Reason,
		InputPayload:   datatypes.JSON(inPayload),
		Candidates:     datatypes.JSON(candidatesPayload),
		ExplainPayload: datatypes.JSON(explainPayload),
	}
	return s.simRepo.Create(ctx, row)
}

func (s *RoutingOptimizerService) ensureDefaultProfile(ctx context.Context) (*LogisticsModel.RoutingScoreProfile, error) {
	row := &LogisticsModel.RoutingScoreProfile{
		ID:               utils.NewUUID(),
		Name:             "default",
		TimelinessWeight: 0.4,
		CostWeight:       0.3,
		QuotaWeight:      0.2,
		RiskWeight:       0.1,
		FallbackStrategy: "highest_timeliness",
		Enabled:          true,
		Config:           datatypes.JSON([]byte("{}")),
	}
	if err := s.profileRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *RoutingOptimizerService) applyFallback(strategy string, rows []RoutingOptimizerCandidate) {
	if len(rows) < 2 {
		return
	}
	switch normalizeFallbackStrategy(strategy) {
	case "lowest_cost":
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Cost == rows[j].Cost {
				return rows[i].CarrierID < rows[j].CarrierID
			}
			return rows[i].Cost < rows[j].Cost
		})
	case "highest_quota":
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Quota == rows[j].Quota {
				return rows[i].CarrierID < rows[j].CarrierID
			}
			return rows[i].Quota > rows[j].Quota
		})
	default:
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Timeliness == rows[j].Timeliness {
				return rows[i].CarrierID < rows[j].CarrierID
			}
			return rows[i].Timeliness > rows[j].Timeliness
		})
	}
}

func sanitizeRoutingWeights(t, c, q, r float64) [4]float64 {
	if t < 0 {
		t = 0
	}
	if c < 0 {
		c = 0
	}
	if q < 0 {
		q = 0
	}
	if r < 0 {
		r = 0
	}
	total := t + c + q + r
	if total <= 0 {
		return [4]float64{0.4, 0.3, 0.2, 0.1}
	}
	return [4]float64{t / total, c / total, q / total, r / total}
}

func normalizeFallbackStrategy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "lowest_cost":
		return "lowest_cost"
	case "highest_quota":
		return "highest_quota"
	default:
		return "highest_timeliness"
	}
}

func filterOptimizerCarriers(rows []LogisticsModel.Carrier, ids []string) []LogisticsModel.Carrier {
	allowed := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}
	result := make([]LogisticsModel.Carrier, 0, len(rows))
	for _, row := range rows {
		if !strings.EqualFold(row.Status, "active") {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[row.ID]; !ok {
				continue
			}
		}
		result = append(result, row)
	}
	return result
}

func parseCarrierMetric(config datatypes.JSON, snake string, camel string, fallback float64) float64 {
	parsed := parseCarrierConfig(config)
	if v, ok := parsed[snake]; ok {
		if n, ok := asFloat(v); ok {
			return n
		}
	}
	if v, ok := parsed[camel]; ok {
		if n, ok := asFloat(v); ok {
			return n
		}
	}
	return fallback
}

func parseCarrierConfig(config datatypes.JSON) map[string]any {
	if len(config) == 0 {
		return map[string]any{}
	}
	var parsed map[string]any
	if err := jsonUnmarshal(config, &parsed); err != nil || parsed == nil {
		return map[string]any{}
	}
	return parsed
}

func jsonUnmarshal(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func round2Optimizer(v float64) float64 {
	return math.Round(v*100) / 100
}

func buildOptimizerExplain(name string, timeliness, cost, quota, risk float64) string {
	return "" + name + " 命中：时效=" + formatRatio(timeliness) + " 成本=" + formatRatio(1-cost) + " 配额=" + formatRatio(quota) + " 风险=" + formatRatio(1-risk)
}

func formatRatio(v float64) string {
	return fmt.Sprintf("%.2f%%", round2Optimizer(v*100))
}
