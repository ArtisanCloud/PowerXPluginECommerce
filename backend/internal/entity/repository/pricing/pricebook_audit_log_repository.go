package repository

import (
	"context"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// PricebookAuditLogRepository wraps BaseRepository with tenant-aware helpers.
type PricebookAuditLogRepository struct {
	*repo.BaseRepository[pricingModel.PricebookAuditLog]
}

func NewPricebookAuditLogRepository(db *gorm.DB) *PricebookAuditLogRepository {
	return &PricebookAuditLogRepository{BaseRepository: repo.NewBaseRepository[pricingModel.PricebookAuditLog](db)}
}

func (r *PricebookAuditLogRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *PricebookAuditLogRepository) Create(ctx context.Context, entry *pricingModel.PricebookAuditLog) (*pricingModel.PricebookAuditLog, error) {
	if entry == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(entry.TenantUUID) == "" {
		entry.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, entry)
}
