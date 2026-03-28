package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LastmileRecoveryService struct {
	ruleRepo *LogisticsRepo.LastmileRecoveryRuleRepository
	runRepo  *LogisticsRepo.LastmileRecoveryRunRepository
}

type UpsertLastmileRecoveryRuleRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	TriggerEvent string         `json:"trigger_event"`
	Action       string         `json:"action"`
	Priority     int            `json:"priority,omitempty"`
	MaxRetries   int            `json:"max_retries,omitempty"`
	Enabled      *bool          `json:"enabled,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type ExecuteLastmileRecoveryRequest struct {
	RequestKey   string `json:"request_key,omitempty"`
	RuleID       string `json:"rule_id,omitempty"`
	WaybillID    string `json:"waybill_id,omitempty"`
	WaybillNo    string `json:"waybill_no,omitempty"`
	TriggerEvent string `json:"trigger_event,omitempty"`
}

type TakeoverLastmileRecoveryRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func NewLastmileRecoveryService(deps *app.Deps) *LastmileRecoveryService {
	if deps == nil || deps.DB == nil {
		return &LastmileRecoveryService{}
	}
	return &LastmileRecoveryService{
		ruleRepo: LogisticsRepo.NewLastmileRecoveryRuleRepository(deps.DB),
		runRepo:  LogisticsRepo.NewLastmileRecoveryRunRepository(deps.DB),
	}
}

func (s *LastmileRecoveryService) ListRules(ctx context.Context, tenantUUID string, enabled *bool) ([]LogisticsModel.LastmileRecoveryRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("lastmile recovery service unavailable")
	}
	return s.ruleRepo.List(withTenantContext(ctx, tenantUUID), enabled)
}

func (s *LastmileRecoveryService) UpsertRule(ctx context.Context, tenantUUID string, req UpsertLastmileRecoveryRuleRequest) (*LogisticsModel.LastmileRecoveryRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("lastmile recovery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	trigger := normalizeLastmileTrigger(req.TriggerEvent)
	action := normalizeLastmileAction(req.Action)
	if trigger == "" || action == "" {
		return nil, errors.New("trigger_event/action is invalid")
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	maxRetries := req.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	cfg, err := jsonBytes(req.Config, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row := &LogisticsModel.LastmileRecoveryRule{
		ID:           strings.TrimSpace(req.ID),
		Name:         name,
		TriggerEvent: trigger,
		Action:       action,
		Priority:     priority,
		MaxRetries:   maxRetries,
		Enabled:      enabled,
		Config:       datatypes.JSON(cfg),
	}
	if row.ID == "" {
		existing, getErr := s.ruleRepo.GetByName(ctx, name)
		if getErr == nil && existing != nil {
			row = existing
			row.TriggerEvent = trigger
			row.Action = action
			row.Priority = priority
			row.MaxRetries = maxRetries
			row.Enabled = enabled
			row.Config = datatypes.JSON(cfg)
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		} else {
			row.ID = utils.NewUUID()
		}
	} else {
		existing, getErr := s.ruleRepo.GetByID(ctx, row.ID)
		if getErr == nil && existing != nil {
			row = existing
			row.Name = name
			row.TriggerEvent = trigger
			row.Action = action
			row.Priority = priority
			row.MaxRetries = maxRetries
			row.Enabled = enabled
			row.Config = datatypes.JSON(cfg)
		}
	}
	if err := s.ruleRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *LastmileRecoveryService) ListRuns(ctx context.Context, tenantUUID, waybillNo, status string, limit int) ([]LogisticsModel.LastmileRecoveryRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("lastmile recovery service unavailable")
	}
	return s.runRepo.List(withTenantContext(ctx, tenantUUID), waybillNo, normalizeLastmileRunStatus(status), limit)
}

func (s *LastmileRecoveryService) Execute(ctx context.Context, tenantUUID string, req ExecuteLastmileRecoveryRequest) (*LogisticsModel.LastmileRecoveryRun, string, error) {
	if s == nil || s.ruleRepo == nil || s.runRepo == nil {
		return nil, "", errors.New("lastmile recovery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "lastmile#" + utils.NewUUID()
	}
	existing, err := s.runRepo.GetByRequestKey(ctx, requestKey)
	if err == nil && existing != nil {
		return existing, "replayed", nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	rule, err := s.pickRule(ctx, req)
	if err != nil {
		return nil, "", err
	}
	row := &LogisticsModel.LastmileRecoveryRun{
		ID:           utils.NewUUID(),
		RequestKey:   requestKey,
		RuleID:       rule.ID,
		WaybillID:    strings.TrimSpace(req.WaybillID),
		WaybillNo:    strings.TrimSpace(req.WaybillNo),
		TriggerEvent: rule.TriggerEvent,
		Action:       rule.Action,
		Status:       "success",
		RetryCount:   0,
		MaxRetries:   rule.MaxRetries,
		Message:      "auto action executed",
		Metadata:     datatypes.JSON([]byte("{}")),
	}
	if strings.TrimSpace(req.TriggerEvent) != "" {
		row.TriggerEvent = normalizeLastmileTrigger(req.TriggerEvent)
		if row.TriggerEvent == "" {
			row.TriggerEvent = rule.TriggerEvent
		}
	}
	if row.TriggerEvent == "lost" || row.TriggerEvent == "timeout" {
		row.Status = "retrying"
		row.RetryCount = 1
		row.Message = "scheduled retry"
	}
	if err := s.runRepo.Save(ctx, row); err != nil {
		return nil, "", err
	}
	return row, "created", nil
}

func (s *LastmileRecoveryService) Takeover(ctx context.Context, tenantUUID, runID string, req TakeoverLastmileRecoveryRequest) (*LogisticsModel.LastmileRecoveryRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("lastmile recovery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.runRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	action := normalizeTakeoverAction(req.Action)
	if action == "" {
		return nil, errors.New("action must be takeover or resume")
	}
	now := time.Now().UTC()
	if action == "takeover" {
		row.ManualTaken = true
		row.TakenBy = strings.TrimSpace(req.OperatorID)
		row.TakenReason = strings.TrimSpace(req.Reason)
		row.TakenAt = &now
		row.Status = "manual_taken"
		row.Message = "manual takeover"
	} else {
		row.ManualTaken = false
		row.TakenBy = strings.TrimSpace(req.OperatorID)
		row.TakenReason = strings.TrimSpace(req.Reason)
		row.Status = "retrying"
		row.Message = "manual resumed automation"
	}
	if err := s.runRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *LastmileRecoveryService) pickRule(ctx context.Context, req ExecuteLastmileRecoveryRequest) (*LogisticsModel.LastmileRecoveryRule, error) {
	if id := strings.TrimSpace(req.RuleID); id != "" {
		return s.ruleRepo.GetByID(ctx, id)
	}
	rows, err := s.ruleRepo.List(ctx, boolPtr(true))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("no enabled lastmile recovery rule found")
	}
	trigger := normalizeLastmileTrigger(req.TriggerEvent)
	if trigger == "" {
		return &rows[0], nil
	}
	for _, row := range rows {
		if row.TriggerEvent == trigger {
			item := row
			return &item, nil
		}
	}
	return &rows[0], nil
}

func normalizeLastmileTrigger(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "delay", "rejected", "lost", "timeout":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeLastmileAction(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "redispatch", "reship", "refund", "manual_review":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeLastmileRunStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pending", "success", "failed", "retrying", "manual_taken":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeTakeoverAction(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "takeover":
		return "takeover"
	case "resume":
		return "resume"
	default:
		return ""
	}
}

func boolPtr(v bool) *bool { return &v }
