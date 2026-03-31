package after_sales

import (
	"context"
	"errors"
	"strings"
	"time"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	afterSalesObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

var ErrDecisionReasonRequired = errors.New("reject reason is required")

type DecisionService struct {
	deps         *app.Deps
	caseService  *CaseService
	decisionRepo *AfterSalesRepo.DecisionRepository
	paymentGuard *PaymentGuardService
	emitter      *afterSalesObs.Emitter
}

func NewDecisionService(deps *app.Deps) *DecisionService {
	if deps == nil || deps.DB == nil {
		return &DecisionService{deps: deps}
	}
	return &DecisionService{
		deps:         deps,
		caseService:  NewCaseService(deps),
		decisionRepo: AfterSalesRepo.NewDecisionRepository(deps.DB),
		paymentGuard: NewPaymentGuardService(deps),
		emitter:      afterSalesObs.NewEmitter(deps.RuntimeLogger(deps.Ctx, "after-sales", nil)),
	}
}

func (s *DecisionService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseService != nil && s.decisionRepo != nil
}

func (s *DecisionService) Approve(ctx context.Context, tenantUUID, caseID, operatorID, note string) (*AfterSalesModel.AfterSaleCase, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	caseDetail, err := s.caseService.Detail(ctx, tenantUUID, caseID)
	if err != nil {
		return nil, err
	}
	if s.paymentGuard != nil {
		if err := s.paymentGuard.EnsureRefundAllowed(ctx, tenantUUID, &AfterSalesModel.AfterSaleCase{
			ID:                   caseDetail.Case.ID,
			OrderID:              caseDetail.Case.OrderID,
			OrderItemID:          caseDetail.Case.OrderItemID,
			CaseType:             caseDetail.Case.CaseType,
			Status:               caseDetail.Case.Status,
			RequestedAmountMinor: caseDetail.Case.RequestedAmountMinor,
			Currency:             caseDetail.Case.Currency,
		}); err != nil {
			return nil, err
		}
	}
	row, err := s.caseService.Transition(ctx, tenantUUID, caseID, "approve", operatorID, note)
	if err != nil {
		return nil, err
	}
	err = s.decisionRepo.Create(withTenantContext(ctx, tenantUUID), &AfterSalesModel.AfterSaleDecision{
		ID:           utils.NewUUID(),
		CaseID:       row.ID,
		Decision:     "approved",
		DecisionNote: strings.TrimSpace(note),
		DecidedBy:    strings.TrimSpace(operatorID),
		DecidedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		reqID, _ := authx.RequestIDFromContext(ctx)
		s.emitter.EmitAudit(afterSalesObs.AuditEvent{
			Action:     "after_sales.approve",
			TenantUUID: strings.TrimSpace(tenantUUID),
			ActorID:    strings.TrimSpace(operatorID),
			TargetID:   strings.TrimSpace(row.ID),
			Result:     "approved",
			Reason:     strings.TrimSpace(note),
			EmittedAt:  time.Now().UTC(),
			Metadata: map[string]any{
				"request_id": reqID,
				"case_no":    row.CaseNo,
				"order_id":   row.OrderID,
			},
		})
	}
	return row, nil
}

func (s *DecisionService) Reject(ctx context.Context, tenantUUID, caseID, operatorID, reasonCode, note string) (*AfterSalesModel.AfterSaleCase, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	if strings.TrimSpace(reasonCode) == "" {
		return nil, ErrDecisionReasonRequired
	}
	row, err := s.caseService.Transition(ctx, tenantUUID, caseID, "reject", operatorID, note)
	if err != nil {
		return nil, err
	}
	err = s.decisionRepo.Create(withTenantContext(ctx, tenantUUID), &AfterSalesModel.AfterSaleDecision{
		ID:               utils.NewUUID(),
		CaseID:           row.ID,
		Decision:         "rejected",
		RejectReasonCode: strings.TrimSpace(reasonCode),
		DecisionNote:     strings.TrimSpace(note),
		DecidedBy:        strings.TrimSpace(operatorID),
		DecidedAt:        time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		reqID, _ := authx.RequestIDFromContext(ctx)
		s.emitter.EmitAudit(afterSalesObs.AuditEvent{
			Action:     "after_sales.reject",
			TenantUUID: strings.TrimSpace(tenantUUID),
			ActorID:    strings.TrimSpace(operatorID),
			TargetID:   strings.TrimSpace(row.ID),
			Result:     "rejected",
			Reason:     strings.TrimSpace(reasonCode),
			EmittedAt:  time.Now().UTC(),
			Metadata: map[string]any{
				"request_id": reqID,
				"case_no":    row.CaseNo,
				"order_id":   row.OrderID,
				"note":       strings.TrimSpace(note),
			},
		})
	}
	return row, nil
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrAdminCaseNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
