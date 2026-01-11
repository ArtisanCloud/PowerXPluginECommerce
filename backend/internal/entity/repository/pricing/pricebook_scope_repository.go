package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// PricebookScopeRepository wraps BaseRepository with tenant-aware helpers.
type PricebookScopeRepository struct {
	*repo.BaseRepository[pricingModel.PricebookScope]
}

func NewPricebookScopeRepository(db *gorm.DB) *PricebookScopeRepository {
	return &PricebookScopeRepository{BaseRepository: repo.NewBaseRepository[pricingModel.PricebookScope](db)}
}

func (r *PricebookScopeRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *PricebookScopeRepository) Create(ctx context.Context, scope *pricingModel.PricebookScope) (*pricingModel.PricebookScope, error) {
	if scope == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(scope.TenantUUID) == "" {
		scope.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, scope)
}
