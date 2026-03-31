package after_sales

import (
	"context"
	"errors"
	"strings"
	"time"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

var ErrDecisionReasonRequired = errors.New("reject reason is required")

type DecisionService struct {
	deps         *app.Deps
	caseService  *CaseService
	decisionRepo *AfterSalesRepo.DecisionRepository
}

func NewDecisionService(deps *app.Deps) *DecisionService {
	if deps == nil || deps.DB == nil {
		return &DecisionService{deps: deps}
	}
	return &DecisionService{
		deps:         deps,
		caseService:  NewCaseService(deps),
		decisionRepo: AfterSalesRepo.NewDecisionRepository(deps.DB),
	}
}

func (s *DecisionService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseService != nil && s.decisionRepo != nil
}

func (s *DecisionService) Approve(ctx context.Context, tenantUUID, caseID, operatorID, note string) (*AfterSalesModel.AfterSaleCase, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
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
	return row, nil
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrAdminCaseNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
