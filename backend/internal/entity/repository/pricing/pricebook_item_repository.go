package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// PricebookItemRepository wraps BaseRepository with tenant-aware helpers.
type PricebookItemRepository struct {
	*repo.BaseRepository[pricingModel.PricebookItem]
}

func NewPricebookItemRepository(db *gorm.DB) *PricebookItemRepository {
	return &PricebookItemRepository{BaseRepository: repo.NewBaseRepository[pricingModel.PricebookItem](db)}
}

func (r *PricebookItemRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *PricebookItemRepository) Create(ctx context.Context, item *pricingModel.PricebookItem) (*pricingModel.PricebookItem, error) {
	if item == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(item.TenantUUID) == "" {
		item.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, item)
}
