package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// PricebookRepository wraps BaseRepository with tenant-aware helpers.
type PricebookRepository struct {
	*repo.BaseRepository[pricingModel.Pricebook]
}

func NewPricebookRepository(db *gorm.DB) *PricebookRepository {
	return &PricebookRepository{BaseRepository: repo.NewBaseRepository[pricingModel.Pricebook](db)}
}

func (r *PricebookRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *PricebookRepository) Create(ctx context.Context, pb *pricingModel.Pricebook) (*pricingModel.Pricebook, error) {
	if pb == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(pb.TenantUUID) == "" {
		pb.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, pb)
}
