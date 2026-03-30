package logistics

import (
	"context"
	"errors"
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

type FulfillmentFinanceRiskService struct {
	riskRepo    *LogisticsRepo.FulfillmentFinanceRiskRepository
	auditRepo   *LogisticsRepo.FulfillmentFinanceRiskAuditRepository
	billingRepo *LogisticsRepo.BillingRepository
	caseRepo    *LogisticsRepo.BillingCaseRepository
}

type FinanceRiskQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	Status    string `json:"status,omitempty"`
	RiskLevel string `json:"risk_level,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type EvaluateFinanceRiskRequest struct {
	CarrierID string  `json:"carrier_id,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
}

type ExecuteFinanceRiskActionRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Note       string `json:"note,omitempty"`
	RequestKey string `json:"request_key,omitempty"`
}

func NewFulfillmentFinanceRiskService(deps *app.Deps) *FulfillmentFinanceRiskService {
	if deps == nil || deps.DB == nil {
		return &FulfillmentFinanceRiskService{}
	}
	waybillRepo := LogisticsRepo.NewWaybillRepository(deps.DB)
	return &FulfillmentFinanceRiskService{
		riskRepo:    LogisticsRepo.NewFulfillmentFinanceRiskRepository(deps.DB),
		auditRepo:   LogisticsRepo.NewFulfillmentFinanceRiskAuditRepository(deps.DB),
		billingRepo: LogisticsRepo.NewBillingRepository(waybillRepo),
		caseRepo:    LogisticsRepo.NewBillingCaseRepository(deps.DB),
	}
}

func (s *FulfillmentFinanceRiskService) ListRisks(ctx context.Context, tenantUUID string, query FinanceRiskQuery) ([]LogisticsModel.FulfillmentFinanceRisk, error) {
	if s == nil || s.riskRepo == nil {
		return nil, errors.New("finance risk service unavailable")
	}
	return s.riskRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.FinanceRiskFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		Status:    normalizeFinanceRiskStatus(query.Status),
		RiskLevel: normalizeFinanceRiskLevel(query.RiskLevel),
		Limit:     query.Limit,
	})
}

func (s *FulfillmentFinanceRiskService) ListAudits(ctx context.Context, tenantUUID, riskID string, limit int) ([]LogisticsModel.FulfillmentFinanceRiskAudit, error) {
	if s == nil || s.auditRepo == nil {
		return nil, errors.New("finance risk service unavailable")
	}
	return s.auditRepo.List(withTenantContext(ctx, tenantUUID), strings.TrimSpace(riskID), limit)
}

func (s *FulfillmentFinanceRiskService) Evaluate(ctx context.Context, tenantUUID string, req EvaluateFinanceRiskRequest) ([]LogisticsModel.FulfillmentFinanceRisk, error) {
	if s == nil || s.riskRepo == nil || s.billingRepo == nil || s.caseRepo == nil {
		return nil, errors.New("finance risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 70
	}
	filter := LogisticsRepo.BillingFilter{CarrierID: strings.TrimSpace(req.CarrierID)}
	items, err := s.billingRepo.ListWaybills(ctx, filter)
	if err != nil {
		return nil, err
	}
	rows := make([]LogisticsModel.FulfillmentFinanceRisk, 0, len(items))
	for _, item := range items {
		payout, chargeback, composite, factors := scoreFinanceRisk(item.EstimatedFee, item.ActualFee, item.DiffFee, item.BillingStatus)
		level := classifyFinanceRiskLevel(composite)
		action := suggestFinanceStopLoss(level)
		status := "open"
		if composite < threshold {
			status = "monitor"
		}
		suggestion := buildFinanceRiskSuggestion(level, action)

		row, err := s.riskRepo.GetByWaybillID(ctx, item.WaybillID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || row == nil {
			row = &LogisticsModel.FulfillmentFinanceRisk{
				ID:        utils.NewUUID(),
				WaybillID: item.WaybillID,
			}
		}
		row.WaybillNo = item.WaybillNo
		row.CarrierID = item.CarrierID
		if billingCase, berr := s.caseRepo.GetActiveByWaybillID(ctx, item.WaybillID); berr == nil && billingCase != nil {
			row.BillingCaseID = billingCase.ID
		}
		row.PayoutRiskScore = payout
		row.ChargebackRiskScore = chargeback
		row.CompositeRiskScore = composite
		row.RiskLevel = level
		row.ThresholdValue = threshold
		row.StopLossAction = action
		row.Status = status
		row.Suggestion = suggestion
		factorsJSON, _ := jsonBytes(factors, []byte("{}"))
		row.RiskFactors = datatypes.JSON(factorsJSON)
		if err := s.riskRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		rows = append(rows, *row)
	}
	return rows, nil
}

func (s *FulfillmentFinanceRiskService) ExecuteAction(ctx context.Context, tenantUUID, riskID string, req ExecuteFinanceRiskActionRequest) (*LogisticsModel.FulfillmentFinanceRisk, string, error) {
	if s == nil || s.riskRepo == nil || s.auditRepo == nil {
		return nil, "", errors.New("finance risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.riskRepo.GetByID(ctx, riskID)
	if err != nil {
		return nil, "", err
	}
	action := normalizeFinanceRiskAction(req.Action)
	if action == "" {
		return nil, "", errors.New("action must be freeze_settlement|hold_payout|manual_review|release|close")
	}
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = strings.TrimSpace(row.ID) + "#" + action
	}
	if _, err := s.auditRepo.GetByRequestKey(ctx, row.ID, requestKey); err == nil {
		return row, "replayed", nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	now := time.Now().UTC()
	row.LastAction = action
	row.LastActionBy = strings.TrimSpace(req.OperatorID)
	row.LastActionAt = &now
	row.ActionCount++

	switch action {
	case "freeze_settlement", "hold_payout", "manual_review":
		row.StopLossAction = action
		row.Status = "mitigating"
	case "release", "close":
		row.StopLossAction = "release"
		row.Status = "closed"
		row.ResolvedAt = &now
	}
	row.Suggestion = buildFinanceRiskSuggestion(row.RiskLevel, row.StopLossAction)
	if err := s.riskRepo.Save(ctx, row); err != nil {
		return nil, "", err
	}
	payloadJSON, _ := jsonBytes(map[string]any{
		"risk_level":            row.RiskLevel,
		"composite_risk_score":  row.CompositeRiskScore,
		"stop_loss_action":      row.StopLossAction,
		"risk_status":           row.Status,
		"waybill_no":            row.WaybillNo,
		"billing_case_id":       row.BillingCaseID,
		"execute_operator_id":   strings.TrimSpace(req.OperatorID),
		"execute_operator_note": strings.TrimSpace(req.Note),
	}, []byte("{}"))
	audit := &LogisticsModel.FulfillmentFinanceRiskAudit{
		ID:         utils.NewUUID(),
		RiskID:     row.ID,
		RequestKey: requestKey,
		Action:     action,
		OperatorID: strings.TrimSpace(req.OperatorID),
		Note:       strings.TrimSpace(req.Note),
		Payload:    datatypes.JSON(payloadJSON),
	}
	if err := s.auditRepo.Save(ctx, audit); err != nil {
		return nil, "", err
	}
	return row, "executed", nil
}

func scoreFinanceRisk(estimatedFee, actualFee, diffFee float64, billingStatus string) (float64, float64, float64, map[string]any) {
	base := math.Max(math.Abs(actualFee), math.Abs(estimatedFee))
	if base < 1 {
		base = 1
	}
	diffRatio := math.Abs(diffFee) / base
	payout := clampFinanceScore(diffRatio*70 + boolScore(diffFee > 20, 20) + boolScore(strings.ToLower(strings.TrimSpace(billingStatus)) != "settled", 10))
	chargeback := clampFinanceScore(diffRatio*45 + boolScore(diffFee < -5, 25) + boolScore(strings.ToLower(strings.TrimSpace(billingStatus)) == "pending", 20))
	composite := roundFinanceRisk2(payout*0.6 + chargeback*0.4)
	return roundFinanceRisk2(payout), roundFinanceRisk2(chargeback), composite, map[string]any{
		"estimated_fee": estimatedFee,
		"actual_fee":    actualFee,
		"diff_fee":      diffFee,
		"diff_ratio":    roundFinanceRisk2(diffRatio * 100),
	}
}

func boolScore(ok bool, score float64) float64 {
	if ok {
		return score
	}
	return 0
}

func clampFinanceScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func roundFinanceRisk2(v float64) float64 {
	return math.Round(v*100) / 100
}

func classifyFinanceRiskLevel(score float64) string {
	switch {
	case score >= 80:
		return "critical"
	case score >= 60:
		return "high"
	case score >= 40:
		return "medium"
	default:
		return "low"
	}
}

func suggestFinanceStopLoss(level string) string {
	switch level {
	case "critical":
		return "freeze_settlement"
	case "high":
		return "hold_payout"
	case "medium":
		return "manual_review"
	default:
		return "observe"
	}
}

func buildFinanceRiskSuggestion(level, action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "freeze_settlement":
		return "建议立即冻结结算并触发人工复核"
	case "hold_payout":
		return "建议暂停赔付并核验拒付证据链"
	case "manual_review":
		return "建议进入人工复核，补充凭证后再结算"
	case "release":
		return "风险已释放，可恢复正常结算"
	}
	switch normalizeFinanceRiskLevel(level) {
	case "critical", "high":
		return "建议执行止损动作并持续监控"
	case "medium":
		return "建议关注波动并降低赔付敞口"
	default:
		return "风险可控，继续观察"
	}
}

func normalizeFinanceRiskStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "open", "monitor", "mitigating", "closed":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeFinanceRiskLevel(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "critical", "high", "medium", "low":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeFinanceRiskAction(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "freeze_settlement", "hold_payout", "manual_review", "release", "close":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}
