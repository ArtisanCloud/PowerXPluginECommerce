package logistics

import (
	"context"
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

type SLOGuardService struct {
	policyRepo *LogisticsRepo.SLOGuardPolicyRepository
	stateRepo  *LogisticsRepo.SLOGuardStateRepository
	gatewaySvc *GatewayMetricsService
}

type UpsertSLOGuardPolicyRequest struct {
	ID                string         `json:"id,omitempty"`
	Name              string         `json:"name"`
	CarrierID         string         `json:"carrier_id,omitempty"`
	WindowHours       int            `json:"window_hours,omitempty"`
	MinSuccessRate    float64        `json:"min_success_rate,omitempty"`
	MaxP95LatencyMS   int            `json:"max_p95_latency_ms,omitempty"`
	MaxFailedRequests int            `json:"max_failed_requests,omitempty"`
	Action            string         `json:"action,omitempty"`
	ThrottleRatio     int            `json:"throttle_ratio,omitempty"`
	Enabled           *bool          `json:"enabled,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

type SLOGuardEvaluateRequest struct {
	WindowHours int    `json:"window_hours,omitempty"`
	CarrierID   string `json:"carrier_id,omitempty"`
}

type SLOGuardReleaseRequest struct {
	PolicyID   string `json:"policy_id"`
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type SLOGuardTriggeredItem struct {
	PolicyID      string  `json:"policy_id"`
	PolicyName    string  `json:"policy_name"`
	CarrierID     string  `json:"carrier_id,omitempty"`
	Action        string  `json:"action"`
	ThrottleRatio int     `json:"throttle_ratio"`
	ReasonCode    string  `json:"reason_code"`
	ReasonMessage string  `json:"reason_message"`
	SuccessRate   float64 `json:"success_rate"`
	FailedCount   int     `json:"failed_count"`
	P95LatencyMS  int     `json:"p95_latency_ms"`
}

type SLOGuardStatusSnapshot struct {
	WindowHours       int                     `json:"window_hours"`
	CarrierID         string                  `json:"carrier_id,omitempty"`
	TotalPolicies     int                     `json:"total_policies"`
	EnabledPolicies   int                     `json:"enabled_policies"`
	ThrottledPolicies int                     `json:"throttled_policies"`
	ThrottleRequired  bool                    `json:"throttle_required"`
	Gateway           GatewayHealthSummary    `json:"gateway"`
	Triggered         []SLOGuardTriggeredItem `json:"triggered"`
}

func NewSLOGuardService(deps *app.Deps) *SLOGuardService {
	if deps == nil || deps.DB == nil {
		return &SLOGuardService{}
	}
	return &SLOGuardService{
		policyRepo: LogisticsRepo.NewSLOGuardPolicyRepository(deps.DB),
		stateRepo:  LogisticsRepo.NewSLOGuardStateRepository(deps.DB),
		gatewaySvc: NewGatewayMetricsService(deps),
	}
}

func (s *SLOGuardService) ListPolicies(ctx context.Context, tenantUUID, carrierID string, enabled *bool, limit int) ([]LogisticsModel.SLOGuardPolicy, error) {
	if s == nil || s.policyRepo == nil {
		return nil, errors.New("slo guard service unavailable")
	}
	return s.policyRepo.List(
		withTenantContext(ctx, tenantUUID),
		LogisticsRepo.SLOGuardPolicyFilter{CarrierID: strings.TrimSpace(carrierID), Enabled: enabled, Limit: limit},
	)
}

func (s *SLOGuardService) UpsertPolicy(ctx context.Context, tenantUUID string, req UpsertSLOGuardPolicyRequest) (*LogisticsModel.SLOGuardPolicy, error) {
	if s == nil || s.policyRepo == nil {
		return nil, errors.New("slo guard service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	row := &LogisticsModel.SLOGuardPolicy{ID: strings.TrimSpace(req.ID)}
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
	meta, err := jsonBytes(req.Metadata, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row.Name = name
	row.CarrierID = strings.TrimSpace(req.CarrierID)
	row.WindowHours = clampInt(req.WindowHours, 1, 24*30, 24)
	row.MinSuccessRate = clampFloat(req.MinSuccessRate, 1, 100, 95)
	row.MaxP95LatencyMS = clampInt(req.MaxP95LatencyMS, 50, 30000, 2000)
	row.MaxFailedRequests = clampInt(req.MaxFailedRequests, 0, 1000000, 10)
	row.Action = normalizeSLOGuardAction(req.Action)
	row.ThrottleRatio = clampInt(req.ThrottleRatio, 1, 100, 50)
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	} else if row.ID != "" && row.CreatedAt.IsZero() {
		row.Enabled = true
	}
	row.Metadata = datatypes.JSON(meta)
	if err := s.policyRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *SLOGuardService) Evaluate(ctx context.Context, tenantUUID string, req SLOGuardEvaluateRequest) (*SLOGuardStatusSnapshot, error) {
	if s == nil || s.policyRepo == nil || s.stateRepo == nil || s.gatewaySvc == nil {
		return nil, errors.New("slo guard service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	policies, err := s.policyRepo.List(ctx, LogisticsRepo.SLOGuardPolicyFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Enabled:   func() *bool { v := true; return &v }(),
		Limit:     200,
	})
	if err != nil {
		return nil, err
	}
	windowHours := clampInt(req.WindowHours, 1, 24*30, 24)
	if len(policies) > 0 && req.WindowHours <= 0 {
		windowHours = policies[0].WindowHours
	}
	gateway, err := s.gatewaySvc.Snapshot(ctx, tenantUUID, windowHours)
	if err != nil {
		return nil, err
	}
	triggered := make([]SLOGuardTriggeredItem, 0)
	for _, policy := range policies {
		hit, code, message := evaluateSLOGuardPolicy(policy, gateway.Summary)
		latest, latestErr := s.stateRepo.GetLatestByPolicyID(ctx, policy.ID)
		if latestErr != nil && !errors.Is(latestErr, gorm.ErrRecordNotFound) {
			return nil, latestErr
		}
		if hit {
			now := time.Now().UTC()
			metrics, mErr := jsonBytes(map[string]any{
				"success_rate":    gateway.Summary.SuccessRate,
				"failed_requests": gateway.Summary.FailedRequests,
				"p95_latency_ms":  gateway.Summary.P95LatencyMS,
				"window_hours":    gateway.Summary.WindowHours,
			}, []byte("{}"))
			if mErr != nil {
				return nil, mErr
			}
			if latest == nil || !strings.EqualFold(strings.TrimSpace(latest.Status), "throttled") {
				if err := s.stateRepo.Save(ctx, &LogisticsModel.SLOGuardState{
					ID:             utils.NewUUID(),
					PolicyID:       policy.ID,
					CarrierID:      policy.CarrierID,
					Status:         "throttled",
					Action:         policy.Action,
					ThrottleRatio:  policy.ThrottleRatio,
					ReasonCode:     code,
					ReasonMessage:  message,
					TriggerMetrics: datatypes.JSON(metrics),
					ActivatedAt:    &now,
				}); err != nil {
					return nil, err
				}
			}
			triggered = append(triggered, SLOGuardTriggeredItem{
				PolicyID:      policy.ID,
				PolicyName:    policy.Name,
				CarrierID:     policy.CarrierID,
				Action:        policy.Action,
				ThrottleRatio: policy.ThrottleRatio,
				ReasonCode:    code,
				ReasonMessage: message,
				SuccessRate:   gateway.Summary.SuccessRate,
				FailedCount:   gateway.Summary.FailedRequests,
				P95LatencyMS:  gateway.Summary.P95LatencyMS,
			})
			continue
		}
		if latest != nil && strings.EqualFold(strings.TrimSpace(latest.Status), "throttled") {
			now := time.Now().UTC()
			if err := s.stateRepo.Save(ctx, &LogisticsModel.SLOGuardState{
				ID:            utils.NewUUID(),
				PolicyID:      policy.ID,
				CarrierID:     policy.CarrierID,
				Status:        "normal",
				Action:        policy.Action,
				ThrottleRatio: 0,
				ReasonCode:    "auto_recovered",
				ReasonMessage: "gateway metrics returned to normal threshold range",
				RecoveredAt:   &now,
			}); err != nil {
				return nil, err
			}
		}
	}
	return &SLOGuardStatusSnapshot{
		WindowHours:       gateway.Summary.WindowHours,
		CarrierID:         strings.TrimSpace(req.CarrierID),
		TotalPolicies:     len(policies),
		EnabledPolicies:   len(policies),
		ThrottledPolicies: len(triggered),
		ThrottleRequired:  len(triggered) > 0,
		Gateway:           gateway.Summary,
		Triggered:         triggered,
	}, nil
}

func (s *SLOGuardService) Status(ctx context.Context, tenantUUID string, req SLOGuardEvaluateRequest) (*SLOGuardStatusSnapshot, error) {
	if s == nil || s.policyRepo == nil || s.stateRepo == nil || s.gatewaySvc == nil {
		return nil, errors.New("slo guard service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	enabled := true
	policies, err := s.policyRepo.List(ctx, LogisticsRepo.SLOGuardPolicyFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Enabled:   &enabled,
		Limit:     200,
	})
	if err != nil {
		return nil, err
	}
	windowHours := clampInt(req.WindowHours, 1, 24*30, 24)
	if len(policies) > 0 && req.WindowHours <= 0 {
		windowHours = policies[0].WindowHours
	}
	gateway, err := s.gatewaySvc.Snapshot(ctx, tenantUUID, windowHours)
	if err != nil {
		return nil, err
	}
	stateRows, err := s.stateRepo.List(ctx, LogisticsRepo.SLOGuardStateFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	latestByPolicy := map[string]LogisticsModel.SLOGuardState{}
	for _, row := range stateRows {
		key := strings.TrimSpace(row.PolicyID)
		if key == "" {
			continue
		}
		if _, exists := latestByPolicy[key]; !exists {
			latestByPolicy[key] = row
		}
	}
	triggered := make([]SLOGuardTriggeredItem, 0)
	for _, policy := range policies {
		row, ok := latestByPolicy[policy.ID]
		if !ok || !strings.EqualFold(strings.TrimSpace(row.Status), "throttled") {
			continue
		}
		triggered = append(triggered, SLOGuardTriggeredItem{
			PolicyID:      policy.ID,
			PolicyName:    policy.Name,
			CarrierID:     policy.CarrierID,
			Action:        policy.Action,
			ThrottleRatio: row.ThrottleRatio,
			ReasonCode:    row.ReasonCode,
			ReasonMessage: row.ReasonMessage,
			SuccessRate:   gateway.Summary.SuccessRate,
			FailedCount:   gateway.Summary.FailedRequests,
			P95LatencyMS:  gateway.Summary.P95LatencyMS,
		})
	}
	return &SLOGuardStatusSnapshot{
		WindowHours:       gateway.Summary.WindowHours,
		CarrierID:         strings.TrimSpace(req.CarrierID),
		TotalPolicies:     len(policies),
		EnabledPolicies:   len(policies),
		ThrottledPolicies: len(triggered),
		ThrottleRequired:  len(triggered) > 0,
		Gateway:           gateway.Summary,
		Triggered:         triggered,
	}, nil
}

func (s *SLOGuardService) ManualRelease(ctx context.Context, tenantUUID string, req SLOGuardReleaseRequest) (*LogisticsModel.SLOGuardState, error) {
	if s == nil || s.policyRepo == nil || s.stateRepo == nil {
		return nil, errors.New("slo guard service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	policyID := strings.TrimSpace(req.PolicyID)
	if policyID == "" {
		return nil, errors.New("policy_id is required")
	}
	policy, err := s.policyRepo.GetByID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	latest, err := s.stateRepo.GetLatestByPolicyID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(latest.Status), "throttled") {
		return nil, errors.New("policy is not throttled")
	}
	now := time.Now().UTC()
	row := &LogisticsModel.SLOGuardState{
		ID:               utils.NewUUID(),
		PolicyID:         policyID,
		CarrierID:        policy.CarrierID,
		Status:           "released",
		Action:           policy.Action,
		ThrottleRatio:    0,
		ReasonCode:       "manual_release",
		ReasonMessage:    strings.TrimSpace(req.Reason),
		RecoveredAt:      &now,
		ManualReleasedBy: strings.TrimSpace(req.OperatorID),
		ManualReason:     strings.TrimSpace(req.Reason),
	}
	if row.ReasonMessage == "" {
		row.ReasonMessage = "manual released by operator"
	}
	if err := s.stateRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func evaluateSLOGuardPolicy(policy LogisticsModel.SLOGuardPolicy, summary GatewayHealthSummary) (bool, string, string) {
	reasons := make([]string, 0, 3)
	if policy.MinSuccessRate > 0 && summary.TotalRequests > 0 && summary.SuccessRate < policy.MinSuccessRate {
		reasons = append(reasons, fmt.Sprintf("success_rate %.2f%% < %.2f%%", summary.SuccessRate, policy.MinSuccessRate))
	}
	if policy.MaxP95LatencyMS > 0 && summary.P95LatencyMS > policy.MaxP95LatencyMS {
		reasons = append(reasons, fmt.Sprintf("p95_latency %dms > %dms", summary.P95LatencyMS, policy.MaxP95LatencyMS))
	}
	if policy.MaxFailedRequests > 0 && summary.FailedRequests > policy.MaxFailedRequests {
		reasons = append(reasons, fmt.Sprintf("failed_requests %d > %d", summary.FailedRequests, policy.MaxFailedRequests))
	}
	if len(reasons) == 0 {
		return false, "", ""
	}
	return true, "threshold_breached", strings.Join(reasons, "; ")
}

func normalizeSLOGuardAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "throttle", "reject":
		return strings.ToLower(strings.TrimSpace(action))
	default:
		return "throttle"
	}
}

func clampInt(value, min, max, fallback int) int {
	if value <= 0 {
		return fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampFloat(value, min, max, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
