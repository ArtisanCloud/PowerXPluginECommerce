package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type CarrierProfileFilter struct {
	Status string
	Limit  int
}

type CarrierProfileRepository struct {
	*Repository[LogisticsModel.CarrierProfile]
}

func NewCarrierProfileRepository(db *gorm.DB) *CarrierProfileRepository {
	return &CarrierProfileRepository{Repository: NewRepository[LogisticsModel.CarrierProfile](db)}
}

func (r *CarrierProfileRepository) List(ctx context.Context, filter CarrierProfileFilter) ([]LogisticsModel.CarrierProfile, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.CarrierProfile
	err = q.Order("composite_score DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CarrierProfileRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.CarrierProfile, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CarrierProfile
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CarrierProfileRepository) GetByCarrierID(ctx context.Context, carrierID string) (*LogisticsModel.CarrierProfile, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CarrierProfile
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND carrier_id = ?", tenantUUID, strings.TrimSpace(carrierID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CarrierProfileRepository) Save(ctx context.Context, row *LogisticsModel.CarrierProfile) error {
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
	return r.DB.WithContext(ctx).Save(row).Error
}
