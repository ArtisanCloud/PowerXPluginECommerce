package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type FulfillmentSandboxScenarioFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	Status          string
	Limit           int
}

type FulfillmentSandboxScenarioRepository struct {
	*Repository[LogisticsModel.FulfillmentSandboxScenario]
}

func NewFulfillmentSandboxScenarioRepository(db *gorm.DB) *FulfillmentSandboxScenarioRepository {
	return &FulfillmentSandboxScenarioRepository{Repository: NewRepository[LogisticsModel.FulfillmentSandboxScenario](db)}
}

func (r *FulfillmentSandboxScenarioRepository) List(ctx context.Context, filter FulfillmentSandboxScenarioFilter) ([]LogisticsModel.FulfillmentSandboxScenario, error) {
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
	var rows []LogisticsModel.FulfillmentSandboxScenario
	err = q.Order("updated_at DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FulfillmentSandboxScenarioRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.FulfillmentSandboxScenario, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentSandboxScenario
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FulfillmentSandboxScenarioRepository) GetByName(ctx context.Context, name string) (*LogisticsModel.FulfillmentSandboxScenario, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentSandboxScenario
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND name = ?", tenantUUID, strings.TrimSpace(name)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FulfillmentSandboxScenarioRepository) Save(ctx context.Context, row *LogisticsModel.FulfillmentSandboxScenario) error {
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

type FulfillmentSandboxRunFilter struct {
	ScenarioID string
	Strategy   string
	Limit      int
}

type FulfillmentSandboxRunRepository struct {
	*Repository[LogisticsModel.FulfillmentSandboxRun]
}

func NewFulfillmentSandboxRunRepository(db *gorm.DB) *FulfillmentSandboxRunRepository {
	return &FulfillmentSandboxRunRepository{Repository: NewRepository[LogisticsModel.FulfillmentSandboxRun](db)}
}

func (r *FulfillmentSandboxRunRepository) List(ctx context.Context, filter FulfillmentSandboxRunFilter) ([]LogisticsModel.FulfillmentSandboxRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.ScenarioID) != "" {
		q = q.Where("scenario_id = ?", strings.TrimSpace(filter.ScenarioID))
	}
	if strings.TrimSpace(filter.Strategy) != "" {
		q = q.Where("strategy = ?", strings.TrimSpace(filter.Strategy))
	}
	var rows []LogisticsModel.FulfillmentSandboxRun
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FulfillmentSandboxRunRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.FulfillmentSandboxRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentSandboxRun
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FulfillmentSandboxRunRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.FulfillmentSandboxRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentSandboxRun
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FulfillmentSandboxRunRepository) Save(ctx context.Context, row *LogisticsModel.FulfillmentSandboxRun) error {
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
