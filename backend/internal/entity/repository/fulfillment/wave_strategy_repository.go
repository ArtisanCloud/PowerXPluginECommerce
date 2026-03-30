package fulfillment

import (
	"context"
	"strings"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	"gorm.io/gorm"
)

// WaveStrategyRepository manages wave strategy persistence.
type WaveStrategyRepository struct {
	*Repository[FulfillmentModel.WaveStrategy]
}

func NewWaveStrategyRepository(db *gorm.DB) *WaveStrategyRepository {
	return &WaveStrategyRepository{Repository: NewRepository[FulfillmentModel.WaveStrategy](db)}
}

func (r *WaveStrategyRepository) List(ctx context.Context, enabledOnly bool) ([]FulfillmentModel.WaveStrategy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var rows []FulfillmentModel.WaveStrategy
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *WaveStrategyRepository) GetByID(ctx context.Context, id string) (*FulfillmentModel.WaveStrategy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.WaveStrategy
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaveStrategyRepository) Create(ctx context.Context, row *FulfillmentModel.WaveStrategy) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(row).Error
}
