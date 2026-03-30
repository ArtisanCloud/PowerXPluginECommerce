package reverse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	ReverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	ReverseRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/reverse"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InspectionService struct {
	waybillRepo    *ReverseRepo.WaybillRepository
	ruleRepo       *ReverseRepo.InspectionRuleRepository
	inspectionRepo *ReverseRepo.WaybillInspectionRepository
}

func NewInspectionService(deps *app.Deps) *InspectionService {
	if deps == nil || deps.DB == nil {
		return &InspectionService{}
	}
	return &InspectionService{
		waybillRepo:    ReverseRepo.NewWaybillRepository(deps.DB),
		ruleRepo:       ReverseRepo.NewInspectionRuleRepository(deps.DB),
		inspectionRepo: ReverseRepo.NewWaybillInspectionRepository(deps.DB),
	}
}

type CreateInspectionRuleRequest struct {
	Name           string         `json:"name"`
	Priority       int            `json:"priority"`
	Condition      map[string]any `json:"condition"`
	Decision       string         `json:"decision"`
	Recommendation string         `json:"recommendation"`
	Notes          string         `json:"notes,omitempty"`
	Enabled        *bool          `json:"enabled,omitempty"`
}

type EvaluateInspectionRequest struct {
	Attributes map[string]any `json:"attributes"`
	OperatorID string         `json:"operator_id,omitempty"`
	Notes      string         `json:"notes,omitempty"`
}

type InspectionDecisionResult struct {
	WaybillID        string                       `json:"waybill_id"`
	Rule             *ReverseModel.InspectionRule `json:"rule,omitempty"`
	Decision         string                       `json:"decision"`
	Recommendation   string                       `json:"recommendation"`
	Reason           string                       `json:"reason"`
	IdempotencyState string                       `json:"idempotency_state"`
}

func (s *InspectionService) ListRules(ctx context.Context, tenantUUID string) ([]ReverseModel.InspectionRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("inspection service unavailable")
	}
	return s.ruleRepo.ListAll(withTenantContext(ctx, tenantUUID))
}

func (s *InspectionService) CreateRule(ctx context.Context, tenantUUID string, req CreateInspectionRuleRequest) (*ReverseModel.InspectionRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("inspection service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	decision := normalizeInspectionDecision(req.Decision)
	if decision == "" {
		return nil, errors.New("decision must be one of: resellable, damaged, repair")
	}
	recommendation := strings.TrimSpace(req.Recommendation)
	if recommendation == "" {
		recommendation = mapDecisionToDisposition(decision)
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	condition, _ := jsonBytes(req.Condition, []byte("{}"))
	row := &ReverseModel.InspectionRule{
		ID:             utils.NewUUID(),
		Name:           name,
		Priority:       priority,
		Enabled:        true,
		ConditionJSON:  datatypes.JSON(condition),
		Decision:       decision,
		Recommendation: recommendation,
		Notes:          strings.TrimSpace(req.Notes),
	}
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	}
	if err := s.ruleRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *InspectionService) EvaluateWaybill(ctx context.Context, tenantUUID, waybillID string, req EvaluateInspectionRequest) (*InspectionDecisionResult, error) {
	if s == nil || s.waybillRepo == nil || s.ruleRepo == nil || s.inspectionRepo == nil {
		return nil, errors.New("inspection service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	waybillID = strings.TrimSpace(waybillID)
	if waybillID == "" {
		return nil, errors.New("waybill id required")
	}
	waybill, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	if waybill.Status != "received" && waybill.Status != "closed" {
		return nil, errors.New("inspection requires received/closed waybill status")
	}
	existed, err := s.inspectionRepo.GetByWaybillID(ctx, waybill.ID)
	if err == nil && existed != nil {
		return s.replayResult(ctx, existed), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rules, err := s.ruleRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	attrs := mergeInspectionAttributes(req.Attributes, waybill.Metadata)
	matchedRule, decision, recommendation, reason := evaluateByRules(rules, attrs)

	payload, _ := jsonBytes(attrs, []byte("{}"))
	record := &ReverseModel.WaybillInspection{
		ID:             utils.NewUUID(),
		WaybillID:      waybill.ID,
		Decision:       decision,
		Recommendation: recommendation,
		Reason:         strings.TrimSpace(reason),
		Attributes:     datatypes.JSON(payload),
	}
	if matchedRule != nil {
		record.RuleID = matchedRule.ID
	}
	if err := s.inspectionRepo.Create(ctx, record); err != nil {
		if isUniqueConstraintError(err) {
			existed, getErr := s.inspectionRepo.GetByWaybillID(ctx, waybill.ID)
			if getErr == nil && existed != nil {
				return s.replayResult(ctx, existed), nil
			}
		}
		return nil, err
	}
	waybill.InspectionResult = decision
	waybill.Disposition = mapDecisionToDisposition(decision)
	if waybill.Status == "received" {
		waybill.Status = "closed"
	}
	if err := s.waybillRepo.Save(ctx, waybill); err != nil {
		return nil, err
	}
	return &InspectionDecisionResult{
		WaybillID:        waybill.ID,
		Rule:             matchedRule,
		Decision:         decision,
		Recommendation:   recommendation,
		Reason:           reason,
		IdempotencyState: "created",
	}, nil
}

func (s *InspectionService) replayResult(ctx context.Context, record *ReverseModel.WaybillInspection) *InspectionDecisionResult {
	var rule *ReverseModel.InspectionRule
	if record == nil {
		return &InspectionDecisionResult{IdempotencyState: "replayed"}
	}
	if strings.TrimSpace(record.RuleID) != "" && s != nil && s.ruleRepo != nil {
		rows, _ := s.ruleRepo.ListAll(ctx)
		for i := range rows {
			if strings.TrimSpace(rows[i].ID) == strings.TrimSpace(record.RuleID) {
				copyRule := rows[i]
				rule = &copyRule
				break
			}
		}
	}
	return &InspectionDecisionResult{
		WaybillID:        record.WaybillID,
		Rule:             rule,
		Decision:         record.Decision,
		Recommendation:   record.Recommendation,
		Reason:           record.Reason,
		IdempotencyState: "replayed",
	}
}

func evaluateByRules(rules []ReverseModel.InspectionRule, attrs map[string]any) (*ReverseModel.InspectionRule, string, string, string) {
	for i := range rules {
		rule := rules[i]
		condition := decodeInspectionCondition(rule.ConditionJSON)
		if matchInspectionCondition(condition, attrs) {
			reason := fmt.Sprintf("matched rule: %s", strings.TrimSpace(rule.Name))
			return &rule, normalizeInspectionDecision(rule.Decision), strings.TrimSpace(rule.Recommendation), reason
		}
	}
	return nil, "repair", "manual_review", "no rule matched"
}

func decodeInspectionCondition(raw []byte) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func matchInspectionCondition(condition map[string]any, attrs map[string]any) bool {
	if len(condition) == 0 {
		return true
	}
	for key, expected := range condition {
		actual, ok := attrs[key]
		if !ok {
			return false
		}
		if !matchInspectionValue(expected, actual) {
			return false
		}
	}
	return true
}

func matchInspectionValue(expected any, actual any) bool {
	switch exp := expected.(type) {
	case string:
		act, ok := actual.(string)
		if !ok {
			return false
		}
		return strings.EqualFold(strings.TrimSpace(exp), strings.TrimSpace(act))
	case bool:
		act, ok := actual.(bool)
		return ok && act == exp
	case float64:
		return toFloat(actual) == exp
	case int:
		return toFloat(actual) == float64(exp)
	case []any:
		for _, option := range exp {
			if matchInspectionValue(option, actual) {
				return true
			}
		}
		return false
	default:
		return fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", actual)
	}
}

func toFloat(v any) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case int32:
		return float64(value)
	default:
		return 0
	}
}

func mergeInspectionAttributes(attrs map[string]any, metadata datatypes.JSON) map[string]any {
	out := map[string]any{}
	for key, value := range attrs {
		out[strings.TrimSpace(key)] = value
	}
	var fromMetadata map[string]any
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &fromMetadata)
	}
	for key, value := range fromMetadata {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = value
		}
	}
	return out
}

func normalizeInspectionDecision(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "resellable":
		return "resellable"
	case "damaged":
		return "damaged"
	case "repair":
		return "repair"
	default:
		return ""
	}
}

func mapDecisionToDisposition(decision string) string {
	switch normalizeInspectionDecision(decision) {
	case "resellable":
		return "restock"
	case "damaged":
		return "compensate"
	default:
		return "repair"
	}
}

func isUniqueConstraintError(err error) bool {
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "unique failed")
}
