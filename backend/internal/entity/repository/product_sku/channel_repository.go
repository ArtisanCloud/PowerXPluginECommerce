package product_sku

import (
	"context"
	"errors"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelRepository persists channel mapping metadata for SKU variants.
type ChannelRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUChannel]
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUChannel](db)}
}

// ListBySKU fetches all channel mappings for a given SKU.
func (r *ChannelRepository) ListBySKU(ctx context.Context, tenantID, skuID string) ([]productskumodel.ProductSKUChannel, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("channel repository is not initialized")
	}
	var rows []productskumodel.ProductSKUChannel
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID).
		Order("channel_code ASC").
		Find(&rows).Error
	return rows, err
}

// UpsertMapping creates or updates a channel mapping using sku_id + channel_code uniqueness.
func (r *ChannelRepository) UpsertMapping(ctx context.Context, mapping *productskumodel.ProductSKUChannel) error {
	if r == nil || r.DB == nil {
		return errors.New("channel repository is not initialized")
	}
	if mapping == nil {
		return errors.New("channel mapping payload is required")
	}
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "sku_id"}, {Name: "channel_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"channel_sku_id",
				"status",
				"sync_mode",
				"publish_time",
				"publish_task_id",
				"last_error",
				"price_override",
				"media_override",
				"metadata",
				"updated_at",
			}),
		}).Create(mapping).Error
}

// FindBySKUAndCode locates a mapping for the given SKU + channel code.
func (r *ChannelRepository) FindBySKUAndCode(ctx context.Context, tenantID, skuID, channelCode string) (*productskumodel.ProductSKUChannel, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("channel repository is not initialized")
	}
	var entity productskumodel.ProductSKUChannel
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND sku_id = ? AND channel_code = ?", tenantID, skuID, channelCode).
		First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}
