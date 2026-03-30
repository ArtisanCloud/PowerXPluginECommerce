package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OpsAutomationService struct {
	policyRepo  *LogisticsRepo.OpsAutomationPolicyRepository
	runRepo     *LogisticsRepo.OpsAutomationRunRepository
	failureRepo *LogisticsRepo.GatewayFailureEventRepository
	stateRepo   *LogisticsRepo.SLOGuardStateRepository
}

type OpsAutomationPolicyQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	Enabled   *bool  `json:"enabled,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type UpsertOpsAutomationPolicyRequest struct {
	ID                     string         `json:"id,omitempty"`
	Name                   string         `json:"name"`
	CarrierID              string         `json:"carrier_id,omitempty"`
	RetryStrategy          map[string]any `json:"retry_strategy,omitempty"`
	CircuitBreakerStrategy map[string]any `json:"circuit_breaker_strategy,omitempty"`
	SuppressionRule        map[string]any `json:"suppression_rule,omitempty"`
	EscalationChain        map[string]any `json:"escalation_chain,omitempty"`
	Enabled                *bool          `json:"enabled,omitempty"`
}

type OpsAutomationRunQuery struct {
	PolicyID  string `json:"policy_id,omitempty"`
	CarrierID string `json:"carrier_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type EvaluateOpsAutomationRequest struct {
	CarrierID     string `json:"carrier_id,omitempty"`
	TriggerSource string `json:"trigger_source,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

type OpsAutomationSnapshot struct {
	SuppressionHits int                               `json:"suppression_hits"`
	AutoRecovered   int                               `json:"auto_recovered"`
	Escalated       int                               `json:"escalated"`
	Runs            []LogisticsModel.OpsAutomationRun `json:"runs"`
}

type OpsAutomationTakeoverRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func NewOpsAutomationService(deps *app.Deps) *OpsAutomationService {
	if deps == nil || deps.DB == nil {
		return &OpsAutomationService{}
	}
	return &OpsAutomationService{
		policyRepo:  LogisticsRepo.NewOpsAutomationPolicyRepository(deps.DB),
		runRepo:     LogisticsRepo.NewOpsAutomationRunRepository(deps.DB),
		failureRepo: LogisticsRepo.NewGatewayFailureEventRepository(deps.DB),
		stateRepo:   LogisticsRepo.NewSLOGuardStateRepository(deps.DB),
	}
}

func (s *OpsAutomationService) ListPolicies(ctx context.Context, tenantUUID string, query OpsAutomationPolicyQuery) ([]LogisticsModel.OpsAutomationPolicy, error) {
	if s == nil || s.policyRepo == nil {
		return nil, errors.New("ops automation service unavailable")
	}
	return s.policyRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.OpsAutomationPolicyFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		Enabled:   query.Enabled,
		Limit:     query.Limit,
	})
}

func (s *OpsAutomationService) UpsertPolicy(ctx context.Context, tenantUUID string, req UpsertOpsAutomationPolicyRequest) (*LogisticsModel.OpsAutomationPolicy, error) {
	if s == nil || s.policyRepo == nil {
		return nil, errors.New("ops automation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	row := &LogisticsModel.OpsAutomationPolicy{ID: strings.TrimSpace(req.ID)}
	if row.ID != "" {
		old, err := s.policyRepo.GetByID(ctx, row.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if old != nil {
			*row = *old
		}
	}
	if strings.TrimSpace(row.ID) == "" {
		row.ID = utils.NewUUID()
	}
	retryJSON, err := jsonBytes(req.RetryStrategy, []byte("{}"))
	if err != nil {
		return nil, err
	}
	cbJSON, err := jsonBytes(req.CircuitBreakerStrategy, []byte("{}"))
	if err != nil {
		return nil, err
	}
	suppressJSON, err := jsonBytes(req.SuppressionRule, []byte("{}"))
	if err != nil {
		return nil, err
	}
	escalationJSON, err := jsonBytes(req.EscalationChain, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row.Name = name
	row.CarrierID = strings.TrimSpace(req.CarrierID)
	row.RetryStrategy = datatypes.JSON(retryJSON)
	row.CircuitBreakerStrategy = datatypes.JSON(cbJSON)
	row.SuppressionRule = datatypes.JSON(suppressJSON)
	row.EscalationChain = datatypes.JSON(escalationJSON)
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	} else if row.CreatedAt.IsZero() {
		row.Enabled = true
	}
	if err := s.policyRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *OpsAutomationService) ListRuns(ctx context.Context, tenantUUID string, query OpsAutomationRunQuery) ([]LogisticsModel.OpsAutomationRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("ops automation service unavailable")
	}
	return s.runRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.OpsAutomationRunFilter{
		PolicyID:  strings.TrimSpace(query.PolicyID),
		CarrierID: strings.TrimSpace(query.CarrierID),
		Status:    strings.TrimSpace(query.Status),
		Limit:     query.Limit,
	})
}

func (s *OpsAutomationService) Evaluate(ctx context.Context, tenantUUID string, req EvaluateOpsAutomationRequest) (*OpsAutomationSnapshot, error) {
	if s == nil || s.policyRepo == nil || s.runRepo == nil || s.failureRepo == nil || s.stateRepo == nil {
		return nil, errors.New("ops automation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	enabled := true
	policies, err := s.policyRepo.List(ctx, LogisticsRepo.OpsAutomationPolicyFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Enabled:   &enabled,
		Limit:     req.Limit,
	})
	if err != nil {
		return nil, err
	}
	failures, err := s.failureRepo.List(ctx, LogisticsRepo.GatewayFailureEventFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Status:    "pending",
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	throttledStates, err := s.stateRepo.List(ctx, LogisticsRepo.SLOGuardStateFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Status:    "throttled",
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	snapshot := &OpsAutomationSnapshot{
		Runs: make([]LogisticsModel.OpsAutomationRun, 0, len(policies)),
	}
	for _, policy := range policies {
		policyFailures := filterFailuresByCarrier(failures, strings.TrimSpace(policy.CarrierID))
		pendingCount := len(policyFailures)
		minFailedEvents := intFromJSON(policy.SuppressionRule, "min_failed_events", 2)
		suppressed := pendingCount < minFailedEvents
		retryEnabled := boolFromJSON(policy.RetryStrategy, "enabled", true)
		maxAutoRecover := intFromJSON(policy.RetryStrategy, "max_attempts", 1)
		throttledCount := countThrottledByCarrier(throttledStates, strings.TrimSpace(policy.CarrierID))
		escalationThreshold := intFromJSON(policy.EscalationChain, "failed_event_threshold", 3)
		requireThrottle := boolFromJSON(policy.EscalationChain, "slo_throttled_required", true)
		autoRecovered := 0

		if !suppressed && retryEnabled && maxAutoRecover > 0 {
			for i := range policyFailures {
				if autoRecovered >= maxAutoRecover {
					break
				}
				if !isAutoRecoverableFailure(policyFailures[i].ErrorClass) {
					continue
				}
				now := time.Now().UTC()
				policyFailures[i].Status = "recovered"
				policyFailures[i].RecoveredAt = &now
				policyFailures[i].NextRetryAt = nil
				policyFailures[i].CircuitOpenTill = nil
				if err := s.failureRepo.Save(ctx, &policyFailures[i]); err != nil {
					return nil, err
				}
				autoRecovered++
			}
			pendingCount -= autoRecovered
			if pendingCount < 0 {
				pendingCount = 0
			}
		}

		escalated := !suppressed && (pendingCount >= escalationThreshold || (requireThrottle && throttledCount > 0))
		status := "observing"
		switch {
		case suppressed:
			status = "suppressed"
		case escalated:
			status = "escalated"
		case autoRecovered > 0:
			status = "auto_recovered"
		}

		now := time.Now().UTC()
		triggerSource := normalizeTriggerSource(req.TriggerSource)
		runPayload, err := jsonBytes(map[string]any{
			"pending_failures":      pendingCount,
			"suppression_threshold": minFailedEvents,
			"auto_recovered":        autoRecovered,
			"throttled_states":      throttledCount,
			"escalation_threshold":  escalationThreshold,
		}, []byte("{}"))
		if err != nil {
			return nil, err
		}
		run := &LogisticsModel.OpsAutomationRun{
			ID:            utils.NewUUID(),
			PolicyID:      policy.ID,
			CarrierID:     strings.TrimSpace(policy.CarrierID),
			TriggerSource: triggerSource,
			TriggerKey:    fmt.Sprintf("%s#%s#%d", policy.ID, triggerSource, now.Unix()),
			Status:        status,
			Suppressed:    suppressed,
			AutoRecovered: autoRecovered > 0,
			Escalated:     escalated,
			Payload:       datatypes.JSON(runPayload),
			StartedAt:     &now,
			FinishedAt:    &now,
		}
		if err := s.runRepo.Save(ctx, run); err != nil {
			return nil, err
		}
		policy.LastEvaluatedAt = &now
		if err := s.policyRepo.Save(ctx, &policy); err != nil {
			return nil, err
		}
		if suppressed {
			snapshot.SuppressionHits++
		}
		if run.AutoRecovered {
			snapshot.AutoRecovered++
		}
		if escalated {
			snapshot.Escalated++
		}
		snapshot.Runs = append(snapshot.Runs, *run)
	}
	return snapshot, nil
}

func (s *OpsAutomationService) Takeover(ctx context.Context, tenantUUID, runID string, req OpsAutomationTakeoverRequest) (*LogisticsModel.OpsAutomationRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("ops automation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(runID) == "" {
		return nil, errors.New("run id is required")
	}
	operatorID := strings.TrimSpace(req.OperatorID)
	if operatorID == "" {
		return nil, errors.New("operator_id is required")
	}
	run, err := s.runRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	action := strings.TrimSpace(strings.ToLower(req.Action))
	if action == "" {
		action = "manual_takeover"
	}
	run.Status = action
	run.TakeoverBy = operatorID
	run.TakeoverReason = strings.TrimSpace(req.Reason)
	run.FinishedAt = &now
	if err := s.runRepo.Save(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func filterFailuresByCarrier(rows []LogisticsModel.GatewayFailureEvent, carrierID string) []LogisticsModel.GatewayFailureEvent {
	if strings.TrimSpace(carrierID) == "" {
		return rows
	}
	filtered := make([]LogisticsModel.GatewayFailureEvent, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.CarrierID) == strings.TrimSpace(carrierID) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func countThrottledByCarrier(rows []LogisticsModel.SLOGuardState, carrierID string) int {
	if strings.TrimSpace(carrierID) == "" {
		return len(rows)
	}
	count := 0
	for _, row := range rows {
		if strings.TrimSpace(row.CarrierID) == strings.TrimSpace(carrierID) {
			count++
		}
	}
	return count
}

func intFromJSON(raw datatypes.JSON, key string, fallback int) int {
	if len(raw) == 0 || strings.TrimSpace(key) == "" {
		return fallback
	}
	m := map[string]any{}
	if err := opsJSONUnmarshal(raw, &m); err != nil {
		return fallback
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	default:
		return fallback
	}
}

func boolFromJSON(raw datatypes.JSON, key string, fallback bool) bool {
	if len(raw) == 0 || strings.TrimSpace(key) == "" {
		return fallback
	}
	m := map[string]any{}
	if err := opsJSONUnmarshal(raw, &m); err != nil {
		return fallback
	}
	v, ok := m[key]
	if !ok {
		return fallback
	}
	switch x := v.(type) {
	case bool:
		return x
	default:
		return fallback
	}
}

func opsJSONUnmarshal(raw []byte, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func isAutoRecoverableFailure(errorClass string) bool {
	switch strings.ToLower(strings.TrimSpace(errorClass)) {
	case "timeout", "server_5xx", "unknown":
		return true
	default:
		return false
	}
}

func normalizeTriggerSource(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "schedule", "alert", "manual_evaluate":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "manual_evaluate"
	}
}
