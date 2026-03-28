package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type CapacityPlanFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	Status          string
	Limit           int
}

type CapacityPlanRepository struct {
	*Repository[LogisticsModel.CapacityPlan]
}

func NewCapacityPlanRepository(db *gorm.DB) *CapacityPlanRepository {
	return &CapacityPlanRepository{Repository: NewRepository[LogisticsModel.CapacityPlan](db)}
}

func (r *CapacityPlanRepository) List(ctx context.Context, filter CapacityPlanFilter) ([]LogisticsModel.CapacityPlan, error) {
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
	var rows []LogisticsModel.CapacityPlan
	err = q.Order("updated_at DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CapacityPlanRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.CapacityPlan, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CapacityPlan
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CapacityPlanRepository) GetByName(ctx context.Context, name string) (*LogisticsModel.CapacityPlan, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CapacityPlan
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND name = ?", tenantUUID, strings.TrimSpace(name)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CapacityPlanRepository) Save(ctx context.Context, row *LogisticsModel.CapacityPlan) error {
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

type AllocationDecisionFilter struct {
	CarrierID string
	WaybillID string
	OrderID   string
	Limit     int
}

type AllocationDecisionRepository struct {
	*Repository[LogisticsModel.AllocationDecision]
}

func NewAllocationDecisionRepository(db *gorm.DB) *AllocationDecisionRepository {
	return &AllocationDecisionRepository{Repository: NewRepository[LogisticsModel.AllocationDecision](db)}
}

func (r *AllocationDecisionRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.AllocationDecision, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.AllocationDecision
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AllocationDecisionRepository) List(ctx context.Context, filter AllocationDecisionFilter) ([]LogisticsModel.AllocationDecision, error) {
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
	if strings.TrimSpace(filter.WaybillID) != "" {
		q = q.Where("waybill_id = ?", strings.TrimSpace(filter.WaybillID))
	}
	if strings.TrimSpace(filter.OrderID) != "" {
		q = q.Where("order_id = ?", strings.TrimSpace(filter.OrderID))
	}
	var rows []LogisticsModel.AllocationDecision
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *AllocationDecisionRepository) Save(ctx context.Context, row *LogisticsModel.AllocationDecision) error {
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
