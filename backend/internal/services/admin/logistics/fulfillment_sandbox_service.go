package logistics

import (
	"context"
	"encoding/json"
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
	"gorm.io/gorm"
)

type FulfillmentSandboxService struct {
	scenarioRepo *LogisticsRepo.FulfillmentSandboxScenarioRepository
	runRepo      *LogisticsRepo.FulfillmentSandboxRunRepository
}

type FulfillmentSandboxScenarioQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Status          string `json:"status,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type UpsertFulfillmentSandboxScenarioRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	CarrierID       string         `json:"carrier_id,omitempty"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	BaselineConfig  map[string]any `json:"baseline_config,omitempty"`
	StrategyConfig  map[string]any `json:"strategy_config,omitempty"`
	Status          string         `json:"status,omitempty"`
	Description     string         `json:"description,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
}

type FulfillmentSandboxRunQuery struct {
	ScenarioID string `json:"scenario_id,omitempty"`
	Strategy   string `json:"strategy,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type RunFulfillmentSandboxRequest struct {
	ScenarioID string `json:"scenario_id"`
	WindowDays int    `json:"window_days,omitempty"`
	Strategy   string `json:"strategy,omitempty"`
	RequestKey string `json:"request_key,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type CompareFulfillmentSandboxRequest struct {
	BaselineRunID  string `json:"baseline_run_id"`
	CandidateRunID string `json:"candidate_run_id"`
}

type FulfillmentSandboxCompareResult struct {
	BaselineRunID     string  `json:"baseline_run_id"`
	CandidateRunID    string  `json:"candidate_run_id"`
	TimelinessDelta   float64 `json:"timeliness_delta"`
	CostDelta         float64 `json:"cost_delta"`
	ExceptionDelta    float64 `json:"exception_delta"`
	RecoveryHourDelta float64 `json:"recovery_hour_delta"`
	ScoreDelta        float64 `json:"score_delta"`
	Recommendation    string  `json:"recommendation"`
}

func NewFulfillmentSandboxService(deps *app.Deps) *FulfillmentSandboxService {
	if deps == nil || deps.DB == nil {
		return &FulfillmentSandboxService{}
	}
	return &FulfillmentSandboxService{
		scenarioRepo: LogisticsRepo.NewFulfillmentSandboxScenarioRepository(deps.DB),
		runRepo:      LogisticsRepo.NewFulfillmentSandboxRunRepository(deps.DB),
	}
}

func (s *FulfillmentSandboxService) ListScenarios(ctx context.Context, tenantUUID string, query FulfillmentSandboxScenarioQuery) ([]LogisticsModel.FulfillmentSandboxScenario, error) {
	if s == nil || s.scenarioRepo == nil {
		return nil, errors.New("fulfillment sandbox service unavailable")
	}
	return s.scenarioRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.FulfillmentSandboxScenarioFilter{
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		Status:          normalizeSandboxScenarioStatus(query.Status),
		Limit:           query.Limit,
	})
}

func (s *FulfillmentSandboxService) UpsertScenario(ctx context.Context, tenantUUID string, req UpsertFulfillmentSandboxScenarioRequest) (*LogisticsModel.FulfillmentSandboxScenario, error) {
	if s == nil || s.scenarioRepo == nil {
		return nil, errors.New("fulfillment sandbox service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	status := normalizeSandboxScenarioStatus(req.Status)
	if status == "" {
		status = "active"
	}
	baselineJSON, err := jsonBytes(req.BaselineConfig, []byte("{}"))
	if err != nil {
		return nil, err
	}
	strategyJSON, err := jsonBytes(req.StrategyConfig, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row := &LogisticsModel.FulfillmentSandboxScenario{ID: strings.TrimSpace(req.ID)}
	if row.ID == "" {
		existing, getErr := s.scenarioRepo.GetByName(ctx, name)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		} else {
			row.ID = utils.NewUUID()
		}
	} else {
		existing, getErr := s.scenarioRepo.GetByID(ctx, row.ID)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
		if strings.TrimSpace(row.ID) == "" {
			row.ID = strings.TrimSpace(req.ID)
		}
	}
	row.Name = name
	row.CarrierID = strings.TrimSpace(req.CarrierID)
	row.WarehouseID = strings.TrimSpace(req.WarehouseID)
	row.DestinationZone = strings.TrimSpace(req.DestinationZone)
	row.BaselineConfig = datatypes.JSON(baselineJSON)
	row.StrategyConfig = datatypes.JSON(strategyJSON)
	row.Status = status
	row.Description = strings.TrimSpace(req.Description)
	if strings.TrimSpace(req.OperatorID) != "" {
		if row.CreatedBy == "" {
			row.CreatedBy = strings.TrimSpace(req.OperatorID)
		}
		row.UpdatedBy = strings.TrimSpace(req.OperatorID)
	}
	if row.ID == "" {
		row.ID = utils.NewUUID()
	}
	if err := s.scenarioRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *FulfillmentSandboxService) ListRuns(ctx context.Context, tenantUUID string, query FulfillmentSandboxRunQuery) ([]LogisticsModel.FulfillmentSandboxRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("fulfillment sandbox service unavailable")
	}
	return s.runRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.FulfillmentSandboxRunFilter{
		ScenarioID: strings.TrimSpace(query.ScenarioID),
		Strategy:   normalizeSandboxStrategy(query.Strategy),
		Limit:      query.Limit,
	})
}

func (s *FulfillmentSandboxService) Run(ctx context.Context, tenantUUID string, req RunFulfillmentSandboxRequest) (*LogisticsModel.FulfillmentSandboxRun, error) {
	if s == nil || s.scenarioRepo == nil || s.runRepo == nil {
		return nil, errors.New("fulfillment sandbox service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	scenarioID := strings.TrimSpace(req.ScenarioID)
	if scenarioID == "" {
		return nil, errors.New("scenario_id is required")
	}
	scenario, err := s.scenarioRepo.GetByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	windowDays := req.WindowDays
	if windowDays <= 0 {
		windowDays = 7
	}
	if windowDays > 90 {
		windowDays = 90
	}
	strategy := normalizeSandboxStrategy(req.Strategy)
	if strategy == "" {
		strategy = "balanced"
	}
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = fmt.Sprintf("sandbox#%s#%s#%d", scenario.ID, strategy, windowDays)
	}
	existing, err := s.runRepo.GetByRequestKey(ctx, requestKey)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	baselineCfg := jsonMapFromRaw(scenario.BaselineConfig)
	strategyCfg := jsonMapFromRaw(scenario.StrategyConfig)

	baseTimeliness := numberFromMap(baselineCfg, "timeliness_rate", 93)
	baseCost := numberFromMap(baselineCfg, "cost_index", 1.0)
	baseException := numberFromMap(baselineCfg, "exception_rate", 3.5)
	baseRecovery := numberFromMap(baselineCfg, "recovery_hours", 12)

	if v, ok := strategyCfg[strategy]; ok {
		if m, okm := v.(map[string]any); okm {
			baseTimeliness = numberFromMap(m, "timeliness_rate", baseTimeliness)
			baseCost = numberFromMap(m, "cost_index", baseCost)
			baseException = numberFromMap(m, "exception_rate", baseException)
			baseRecovery = numberFromMap(m, "recovery_hours", baseRecovery)
		}
	}

	factor := 1.0 + (float64(windowDays)-7.0)*0.01
	timeliness := clampSandboxRate(baseTimeliness-(factor-1)*1.8, 70, 99.9)
	cost := math.Max(0.1, baseCost*(1+(factor-1)*0.12))
	exceptionRate := clampSandboxRate(baseException*(1+(factor-1)*0.25), 0.1, 30)
	recoveryHours := math.Max(0.1, baseRecovery*(1+(factor-1)*0.3))

	switch strategy {
	case "timeliness_first":
		timeliness = clampSandboxRate(timeliness+2.5, 70, 99.9)
		cost = cost * 1.12
		exceptionRate = clampSandboxRate(exceptionRate*0.88, 0.1, 30)
	case "cost_first":
		timeliness = clampSandboxRate(timeliness-1.8, 70, 99.9)
		cost = cost * 0.88
		exceptionRate = clampSandboxRate(exceptionRate*1.15, 0.1, 30)
	case "resilience_first":
		timeliness = clampSandboxRate(timeliness-0.6, 70, 99.9)
		cost = cost * 1.05
		exceptionRate = clampSandboxRate(exceptionRate*0.75, 0.1, 30)
		recoveryHours = math.Max(0.1, recoveryHours*0.7)
	default:
	}

	score := roundSandbox2(
		0.45*timeliness +
			0.25*(100/(1+cost)) +
			0.20*(100-exceptionRate*2) +
			0.10*(100-math.Min(recoveryHours*2, 100)),
	)
	recommendation := buildSandboxRecommendation(strategy, timeliness, cost, exceptionRate, recoveryHours)
	snapshot, _ := jsonBytes(map[string]any{
		"timeliness_rate": timeliness,
		"cost_index":      roundSandbox2(cost),
		"exception_rate":  exceptionRate,
		"recovery_hours":  roundSandbox2(recoveryHours),
		"window_days":     windowDays,
		"strategy":        strategy,
	}, []byte("{}"))

	row := &LogisticsModel.FulfillmentSandboxRun{
		ID:             utils.NewUUID(),
		ScenarioID:     scenario.ID,
		RequestKey:     requestKey,
		WindowDays:     windowDays,
		Strategy:       strategy,
		TimelinessRate: roundSandbox2(timeliness),
		CostIndex:      roundSandbox2(cost),
		ExceptionRate:  roundSandbox2(exceptionRate),
		RecoveryHours:  roundSandbox2(recoveryHours),
		Score:          score,
		Snapshot:       datatypes.JSON(snapshot),
		Recommendation: recommendation,
		Status:         "completed",
		CreatedBy:      strings.TrimSpace(req.OperatorID),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.runRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *FulfillmentSandboxService) Compare(ctx context.Context, tenantUUID string, req CompareFulfillmentSandboxRequest) (*FulfillmentSandboxCompareResult, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("fulfillment sandbox service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	baseID := strings.TrimSpace(req.BaselineRunID)
	candidateID := strings.TrimSpace(req.CandidateRunID)
	if baseID == "" || candidateID == "" {
		return nil, errors.New("baseline_run_id and candidate_run_id are required")
	}
	baseRun, err := s.runRepo.GetByID(ctx, baseID)
	if err != nil {
		return nil, err
	}
	candidateRun, err := s.runRepo.GetByID(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	result := &FulfillmentSandboxCompareResult{
		BaselineRunID:     baseRun.ID,
		CandidateRunID:    candidateRun.ID,
		TimelinessDelta:   roundSandbox2(candidateRun.TimelinessRate - baseRun.TimelinessRate),
		CostDelta:         roundSandbox2(candidateRun.CostIndex - baseRun.CostIndex),
		ExceptionDelta:    roundSandbox2(candidateRun.ExceptionRate - baseRun.ExceptionRate),
		RecoveryHourDelta: roundSandbox2(candidateRun.RecoveryHours - baseRun.RecoveryHours),
		ScoreDelta:        roundSandbox2(candidateRun.Score - baseRun.Score),
	}
	if result.ScoreDelta >= 0 {
		result.Recommendation = fmt.Sprintf("推荐候选策略 %s（综合评分 %+0.2f）", candidateRun.Strategy, result.ScoreDelta)
	} else {
		result.Recommendation = fmt.Sprintf("建议保留基线策略 %s（候选评分 %+0.2f）", baseRun.Strategy, result.ScoreDelta)
	}
	return result, nil
}

func normalizeSandboxScenarioStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "inactive", "archived":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return ""
	}
}

func normalizeSandboxStrategy(strategy string) string {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "balanced", "timeliness_first", "cost_first", "resilience_first":
		return strings.ToLower(strings.TrimSpace(strategy))
	default:
		return ""
	}
}

func numberFromMap(raw map[string]any, key string, fallback float64) float64 {
	if raw == nil {
		return fallback
	}
	val, ok := raw[key]
	if !ok || val == nil {
		return fallback
	}
	switch t := val.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case string:
		var out float64
		_, err := fmt.Sscanf(strings.TrimSpace(t), "%f", &out)
		if err == nil {
			return out
		}
	}
	return fallback
}

func jsonMapFromRaw(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	tmp := map[string]any{}
	if err := json.Unmarshal(raw, &tmp); err == nil {
		return tmp
	}
	return map[string]any{}
}

func buildSandboxRecommendation(strategy string, timeliness, cost, exceptionRate, recoveryHours float64) string {
	if exceptionRate > 8 || recoveryHours > 20 {
		return "建议优先提升韧性策略并增加异常恢复资源"
	}
	if timeliness < 90 {
		return "建议提升时效优先权重并优化承运商路由"
	}
	if cost > 1.2 {
		return "建议切换成本优先并控制高成本履约路径"
	}
	return fmt.Sprintf("当前策略 %s 表现稳定，可作为推荐基线", strategy)
}

func clampSandboxRate(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func roundSandbox2(v float64) float64 {
	return math.Round(v*100) / 100
}
