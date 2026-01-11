package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// LockByID loads the pricebook row with an UPDATE lock when supported by the dialect.
func (r *PricebookRepository) LockByID(ctx context.Context, pricebookID string) (*pricingModel.Pricebook, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.DB.WithContext(ctx)
	if r.DB != nil && r.DB.Dialector != nil && r.DB.Dialector.Name() != "sqlite" {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var pb pricingModel.Pricebook
	if err := db.Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(pricebookID)).First(&pb).Error; err != nil {
		return nil, err
	}
	return &pb, nil
}
