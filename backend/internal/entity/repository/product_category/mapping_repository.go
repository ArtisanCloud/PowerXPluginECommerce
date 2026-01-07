package product_category

import (
	"context"
	"errors"
	"strings"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MappingRepository persists channel mappings for categories.
type MappingRepository struct {
	*repository.BaseRepository[productcategory.CategoryMapping]
}

func NewMappingRepository(db *gorm.DB) *MappingRepository {
	return &MappingRepository{
		BaseRepository: repository.NewBaseRepository[productcategory.CategoryMapping](db),
	}
}

func (r *MappingRepository) ListByCategory(ctx context.Context, tenantUUID string, categoryID string) ([]productcategory.CategoryMapping, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("mapping repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	categoryID = strings.TrimSpace(categoryID)
	if tenantUUID == "" || categoryID == "" {
		return nil, errors.New("tenant_uuid and category_id are required")
	}
	var records []productcategory.CategoryMapping
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryMapping{}).
		Where("tenant_uuid = ? AND category_id = ?", tenantUUID, categoryID).
		Order("channel ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *MappingRepository) Upsert(ctx context.Context, tx *gorm.DB, mapping *productcategory.CategoryMapping) (*productcategory.CategoryMapping, error) {
	if r == nil {
		return nil, errors.New("mapping repository not initialized")
	}
	if mapping == nil {
		return nil, errors.New("mapping is required")
	}
	if tx == nil {
		tx = r.DB
	}
	if tx == nil {
		return nil, errors.New("mapping repository database is not initialized")
	}
	mapping.Normalize()
	if mapping.TenantUUID == "" || mapping.CategoryID == "" || mapping.Channel == "" {
		return nil, errors.New("tenant_uuid, category_id and channel are required")
	}
	if err := tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "tenant_uuid"},
				{Name: "category_id"},
				{Name: "channel"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"platform_category_id",
				"strategy",
				"sync_status",
				"metadata",
				"updated_at",
				"deleted_at",
			}),
		}).
		Create(mapping).Error; err != nil {
		return nil, err
	}
	return mapping, nil
}

func (r *MappingRepository) DeleteByCategoryChannel(ctx context.Context, tenantUUID string, categoryID string, channel string) error {
	if r == nil || r.DB == nil {
		return errors.New("mapping repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	categoryID = strings.TrimSpace(categoryID)
	channel = strings.TrimSpace(channel)
	if tenantUUID == "" || categoryID == "" || channel == "" {
		return errors.New("tenant_uuid, category_id and channel are required")
	}
	return r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND category_id = ? AND channel = ?", tenantUUID, categoryID, channel).
		Delete(&productcategory.CategoryMapping{}).Error
}
