package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type InterwarehouseAllocationFilter struct {
	RequestKey        string
	CarrierID         string
	SourceWarehouseID string
	TargetWarehouseID string
	Status            string
	Limit             int
}

type InterwarehouseAllocationRepository struct {
	*Repository[LogisticsModel.InterwarehouseAllocation]
}

func NewInterwarehouseAllocationRepository(db *gorm.DB) *InterwarehouseAllocationRepository {
	return &InterwarehouseAllocationRepository{Repository: NewRepository[LogisticsModel.InterwarehouseAllocation](db)}
}

func (r *InterwarehouseAllocationRepository) List(ctx context.Context, filter InterwarehouseAllocationFilter) ([]LogisticsModel.InterwarehouseAllocation, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.RequestKey) != "" {
		q = q.Where("request_key = ?", strings.TrimSpace(filter.RequestKey))
	}
	if strings.TrimSpace(filter.CarrierID) != "" {
		q = q.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.SourceWarehouseID) != "" {
		q = q.Where("source_warehouse_id = ?", strings.TrimSpace(filter.SourceWarehouseID))
	}
	if strings.TrimSpace(filter.TargetWarehouseID) != "" {
		q = q.Where("target_warehouse_id = ?", strings.TrimSpace(filter.TargetWarehouseID))
	}
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.InterwarehouseAllocation
	err = q.Order("score DESC, target_available DESC, target_warehouse_id ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *InterwarehouseAllocationRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.InterwarehouseAllocation, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.InterwarehouseAllocation
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *InterwarehouseAllocationRepository) Save(ctx context.Context, row *LogisticsModel.InterwarehouseAllocation) error {
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

func (r *InterwarehouseAllocationRepository) SaveBatch(ctx context.Context, rows []LogisticsModel.InterwarehouseAllocation) error {
	if len(rows) == 0 {
		return nil
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].TenantUUID) == "" {
			rows[i].TenantUUID = tenantUUID
		}
	}
	return r.DB.WithContext(ctx).Save(&rows).Error
}
