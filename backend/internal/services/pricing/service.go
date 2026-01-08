package pricing

import (
	"context"
	"errors"
	"strings"

	pricingRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// Service is the pricing/pricebook domain orchestrator.
type Service struct {
	deps                 *app.Deps
	PricebookRepo        *pricingRepo.PricebookRepository
	PricebookVersionRepo *pricingRepo.PricebookVersionRepository
	PricebookScopeRepo   *pricingRepo.PricebookScopeRepository
	PricebookItemRepo    *pricingRepo.PricebookItemRepository
	AuditLogRepo         *pricingRepo.PricebookAuditLogRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	return &Service{
		deps:                 deps,
		PricebookRepo:        pricingRepo.NewPricebookRepository(deps.DB),
		PricebookVersionRepo: pricingRepo.NewPricebookVersionRepository(deps.DB),
		PricebookScopeRepo:   pricingRepo.NewPricebookScopeRepository(deps.DB),
		PricebookItemRepo:    pricingRepo.NewPricebookItemRepository(deps.DB),
		AuditLogRepo:         pricingRepo.NewPricebookAuditLogRepository(deps.DB),
	}
}

func (s *Service) Deps() *app.Deps {
	if s == nil {
		return nil
	}
	return s.deps
}

func (s *Service) Ready() bool { return s != nil && s.deps != nil && s.deps.DB != nil }

func (s *Service) HealthProbe(ctx context.Context) error {
	if !s.Ready() {
		return errors.New("pricing service not ready")
	}
	return ctx.Err()
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("request context missing")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", errors.New("tenant context missing")
}
