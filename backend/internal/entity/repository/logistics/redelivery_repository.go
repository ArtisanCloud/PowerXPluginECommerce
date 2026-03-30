package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// RedeliveryTaskRepository manages redelivery task persistence.
type RedeliveryTaskRepository struct {
	*Repository[LogisticsModel.RedeliveryTask]
}

func NewRedeliveryTaskRepository(db *gorm.DB) *RedeliveryTaskRepository {
	return &RedeliveryTaskRepository{Repository: NewRepository[LogisticsModel.RedeliveryTask](db)}
}

func (r *RedeliveryTaskRepository) List(ctx context.Context, waybillID, status string) ([]LogisticsModel.RedeliveryTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillID) != "" {
		query = query.Where("waybill_id = ?", strings.TrimSpace(waybillID))
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.RedeliveryTask
	err = query.Order("updated_at DESC").Find(&rows).Error
	return rows, err
}

func (r *RedeliveryTaskRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.RedeliveryTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RedeliveryTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RedeliveryTaskRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.RedeliveryTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RedeliveryTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RedeliveryTaskRepository) Create(ctx context.Context, row *LogisticsModel.RedeliveryTask) error {
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

func (r *RedeliveryTaskRepository) Save(ctx context.Context, row *LogisticsModel.RedeliveryTask) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
