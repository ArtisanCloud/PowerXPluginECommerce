package product_category

import (
	"context"
	"errors"
	"strings"

	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_category"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

// Service hosts category/template/mapping orchestration for admin + miniapp wiring.
// Concrete business methods will be implemented in later phases.
type Service struct {
	deps *app.Deps

	CategoryRepo *repo.CategoryRepository
	TemplateRepo *repo.TemplateRepository
	MappingRepo  *repo.MappingRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	return &Service{
		deps:         deps,
		CategoryRepo: repo.NewCategoryRepository(deps.DB),
		TemplateRepo: repo.NewTemplateRepository(deps.DB),
		MappingRepo:  repo.NewMappingRepository(deps.DB),
	}
}

func (s *Service) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil
}

func (s *Service) DB() *gorm.DB {
	if s == nil || s.deps == nil {
		return nil
	}
	return s.deps.DB
}

func (s *Service) TenantUUID(ctx context.Context) (string, error) {
	return s.tenantFromContext(ctx)
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("request context missing")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return strings.TrimSpace(tid), nil
	}
	return "", errors.New("tenant context missing")
}
