package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// PricebookVersionRepository wraps BaseRepository with tenant-aware helpers.
type PricebookVersionRepository struct {
	*repo.BaseRepository[pricingModel.PricebookVersion]
}

func NewPricebookVersionRepository(db *gorm.DB) *PricebookVersionRepository {
	return &PricebookVersionRepository{BaseRepository: repo.NewBaseRepository[pricingModel.PricebookVersion](db)}
}

func (r *PricebookVersionRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *PricebookVersionRepository) Create(ctx context.Context, v *pricingModel.PricebookVersion) (*pricingModel.PricebookVersion, error) {
	if v == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(v.TenantUUID) == "" {
		v.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, v)
}
