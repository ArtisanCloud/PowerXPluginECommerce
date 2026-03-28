package logistics

import (
	"context"
	"errors"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type ExceptionOrchestrationService struct {
	ruleRepo    *LogisticsRepo.ExceptionOrchestrationRuleRepository
	runRepo     *LogisticsRepo.ExceptionOrchestrationRunRepository
	waybillRepo *LogisticsRepo.WaybillRepository
}

type UpsertExceptionRuleRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	TriggerEvent string         `json:"trigger_event"`
	Action       string         `json:"action"`
	Priority     int            `json:"priority,omitempty"`
	Enabled      *bool          `json:"enabled,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type ExecuteExceptionRuleRequest struct {
	RuleID    string `json:"rule_id"`
	WaybillID string `json:"waybill_id,omitempty"`
	WaybillNo string `json:"waybill_no,omitempty"`
	Trigger   string `json:"trigger,omitempty"`
}

func NewExceptionOrchestrationService(deps *app.Deps) *ExceptionOrchestrationService {
	if deps == nil || deps.DB == nil {
		return &ExceptionOrchestrationService{}
	}
	return &ExceptionOrchestrationService{
		ruleRepo:    LogisticsRepo.NewExceptionOrchestrationRuleRepository(deps.DB),
		runRepo:     LogisticsRepo.NewExceptionOrchestrationRunRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
	}
}

func (s *ExceptionOrchestrationService) ListRules(ctx context.Context, tenantUUID string, enabled *bool) ([]LogisticsModel.ExceptionOrchestrationRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("exception orchestration service unavailable")
	}
	return s.ruleRepo.List(withTenantContext(ctx, tenantUUID), enabled)
}

func (s *ExceptionOrchestrationService) UpsertRule(ctx context.Context, tenantUUID string, req UpsertExceptionRuleRequest) (*LogisticsModel.ExceptionOrchestrationRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("exception orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	trigger := strings.TrimSpace(req.TriggerEvent)
	action := strings.TrimSpace(req.Action)
	if name == "" || trigger == "" || action == "" {
		return nil, errors.New("name/trigger_event/action required")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	if strings.TrimSpace(req.ID) != "" {
		row, err := s.ruleRepo.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		cfg, _ := jsonBytes(req.Config, []byte("{}"))
		row.Name = name
		row.TriggerEvent = trigger
		row.Action = action
		row.Priority = priority
		row.Enabled = enabled
		row.Config = datatypes.JSON(cfg)
		if err := s.ruleRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	cfg, _ := jsonBytes(req.Config, []byte("{}"))
	row := &LogisticsModel.ExceptionOrchestrationRule{
		ID:           utils.NewUUID(),
		Name:         name,
		TriggerEvent: trigger,
		Action:       action,
		Priority:     priority,
		Enabled:      enabled,
		Config:       datatypes.JSON(cfg),
	}
	if err := s.ruleRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ExceptionOrchestrationService) ListRuns(ctx context.Context, tenantUUID, waybillNo string, limit int) ([]LogisticsModel.ExceptionOrchestrationRun, error) {
	if s == nil || s.runRepo == nil {
		return nil, errors.New("exception orchestration service unavailable")
	}
	return s.runRepo.List(withTenantContext(ctx, tenantUUID), waybillNo, limit)
}

func (s *ExceptionOrchestrationService) Execute(ctx context.Context, tenantUUID string, req ExecuteExceptionRuleRequest) (*LogisticsModel.ExceptionOrchestrationRun, error) {
	if s == nil || s.ruleRepo == nil || s.runRepo == nil {
		return nil, errors.New("exception orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	ruleID := strings.TrimSpace(req.RuleID)
	if ruleID == "" {
		return nil, errors.New("rule_id required")
	}
	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	waybillID := strings.TrimSpace(req.WaybillID)
	waybillNo := strings.TrimSpace(req.WaybillNo)
	if waybillID == "" && waybillNo != "" && s.waybillRepo != nil {
		if wb, wbErr := s.waybillRepo.GetByWaybillNo(ctx, waybillNo); wbErr == nil && wb != nil {
			waybillID = wb.ID
			waybillNo = wb.WaybillNo
		}
	}
	if waybillNo == "" && waybillID != "" && s.waybillRepo != nil {
		if wb, wbErr := s.waybillRepo.GetByID(ctx, waybillID); wbErr == nil && wb != nil {
			waybillNo = wb.WaybillNo
		}
	}
	if waybillNo == "" {
		return nil, errors.New("waybill_id or waybill_no required")
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = rule.TriggerEvent
	}
	result := "success"
	message := "orchestrated"
	switch strings.ToLower(strings.TrimSpace(rule.Action)) {
	case "auto_compensate":
		message = "triggered auto compensation"
	case "create_ticket":
		message = "created exception ticket"
	case "escalate":
		message = "sla escalated"
	default:
		result = "ignored"
		message = "unsupported action fallback to manual queue"
	}
	run := &LogisticsModel.ExceptionOrchestrationRun{
		ID:        utils.NewUUID(),
		RuleID:    rule.ID,
		WaybillID: waybillID,
		WaybillNo: waybillNo,
		Trigger:   trigger,
		Result:    result,
		Message:   message,
		Metadata:  datatypes.JSON([]byte("{}")),
	}
	if err := s.runRepo.Create(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}
