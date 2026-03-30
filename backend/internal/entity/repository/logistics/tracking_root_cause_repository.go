package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type TrackingRootCauseFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	AnomalyType     string
	Status          string
	Limit           int
}

type TrackingRootCauseRepository struct {
	*Repository[LogisticsModel.TrackingRootCause]
}

func NewTrackingRootCauseRepository(db *gorm.DB) *TrackingRootCauseRepository {
	return &TrackingRootCauseRepository{Repository: NewRepository[LogisticsModel.TrackingRootCause](db)}
}

func (r *TrackingRootCauseRepository) List(ctx context.Context, filter TrackingRootCauseFilter) ([]LogisticsModel.TrackingRootCause, error) {
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
	if strings.TrimSpace(filter.AnomalyType) != "" {
		q = q.Where("anomaly_type = ?", strings.TrimSpace(filter.AnomalyType))
	}
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.TrackingRootCause
	err = q.Order("created_at DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *TrackingRootCauseRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.TrackingRootCause, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.TrackingRootCause
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TrackingRootCauseRepository) GetByWaybillAndAnomaly(ctx context.Context, waybillID, anomalyType string) (*LogisticsModel.TrackingRootCause, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.TrackingRootCause
	err = r.DB.WithContext(ctx).
		Where(
			"tenant_uuid = ? AND waybill_id = ? AND anomaly_type = ?",
			tenantUUID,
			strings.TrimSpace(waybillID),
			strings.TrimSpace(anomalyType),
		).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TrackingRootCauseRepository) Save(ctx context.Context, row *LogisticsModel.TrackingRootCause) error {
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
