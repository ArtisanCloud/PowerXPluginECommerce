package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type CapacityForecastFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	Status          string
	Limit           int
}

type CapacityForecastRepository struct {
	*Repository[LogisticsModel.CapacityForecast]
}

func NewCapacityForecastRepository(db *gorm.DB) *CapacityForecastRepository {
	return &CapacityForecastRepository{Repository: NewRepository[LogisticsModel.CapacityForecast](db)}
}

func (r *CapacityForecastRepository) List(ctx context.Context, filter CapacityForecastFilter) ([]LogisticsModel.CapacityForecast, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		q = q.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.WarehouseID) != "" {
		q = q.Where("warehouse_id = ?", strings.TrimSpace(filter.WarehouseID))
	}
	if strings.TrimSpace(filter.DestinationZone) != "" {
		q = q.Where("destination_zone = ?", strings.TrimSpace(filter.DestinationZone))
	}
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.CapacityForecast
	err = q.Order("created_at DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CapacityForecastRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.CapacityForecast, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CapacityForecast
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CapacityForecastRepository) Save(ctx context.Context, row *LogisticsModel.CapacityForecast) error {
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
