package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type ControlTowerSnapshotFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	WindowHours     int
	Limit           int
}

type ControlTowerSnapshotRepository struct {
	*Repository[LogisticsModel.ControlTowerSnapshot]
}

func NewControlTowerSnapshotRepository(db *gorm.DB) *ControlTowerSnapshotRepository {
	return &ControlTowerSnapshotRepository{Repository: NewRepository[LogisticsModel.ControlTowerSnapshot](db)}
}

func (r *ControlTowerSnapshotRepository) Create(ctx context.Context, row *LogisticsModel.ControlTowerSnapshot) error {
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

func (r *ControlTowerSnapshotRepository) List(ctx context.Context, filter ControlTowerSnapshotFilter) ([]LogisticsModel.ControlTowerSnapshot, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
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
	if filter.WindowHours > 0 {
		q = q.Where("window_hours = ?", filter.WindowHours)
	}
	var rows []LogisticsModel.ControlTowerSnapshot
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

type ControlTowerAlertSubscriptionRepository struct {
	*Repository[LogisticsModel.ControlTowerAlertSubscription]
}

func NewControlTowerAlertSubscriptionRepository(db *gorm.DB) *ControlTowerAlertSubscriptionRepository {
	return &ControlTowerAlertSubscriptionRepository{
		Repository: NewRepository[LogisticsModel.ControlTowerAlertSubscription](db),
	}
}

func (r *ControlTowerAlertSubscriptionRepository) List(ctx context.Context, enabled *bool) ([]LogisticsModel.ControlTowerAlertSubscription, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabled != nil {
		q = q.Where("enabled = ?", *enabled)
	}
	var rows []LogisticsModel.ControlTowerAlertSubscription
	err = q.Order("updated_at DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *ControlTowerAlertSubscriptionRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ControlTowerAlertSubscription, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ControlTowerAlertSubscription
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ControlTowerAlertSubscriptionRepository) Save(ctx context.Context, row *LogisticsModel.ControlTowerAlertSubscription) error {
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
