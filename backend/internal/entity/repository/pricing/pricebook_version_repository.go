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

func (r *PricebookVersionRepository) MaxVersion(ctx context.Context, pricebookID string) (int, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return 0, err
	}
	var maxVer int
	if err := r.DB.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ?", tenantUUID, strings.TrimSpace(pricebookID)).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVer).Error; err != nil {
		return 0, err
	}
	return maxVer, nil
}

// LockActiveVersions applies an UPDATE lock on active versions (Postgres) to serialize publish operations.
func (r *PricebookVersionRepository) LockActiveVersions(ctx context.Context, pricebookID string) error {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	db := r.DB.WithContext(ctx)
	if r.DB != nil && r.DB.Dialector != nil && r.DB.Dialector.Name() != "sqlite" {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var ids []string
	return db.Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ? AND state = ?", tenantUUID, strings.TrimSpace(pricebookID), "active").
		Pluck("id", &ids).Error
}
