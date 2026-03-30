package fulfillment

import (
	"context"
	"strings"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	"gorm.io/gorm"
)

type OutboundRepository struct {
	*Repository[FulfillmentModel.Outbound]
}

func NewOutboundRepository(db *gorm.DB) *OutboundRepository {
	return &OutboundRepository{Repository: NewRepository[FulfillmentModel.Outbound](db)}
}

func (r *OutboundRepository) Create(ctx context.Context, row *FulfillmentModel.Outbound) error {
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

func (r *OutboundRepository) GetByID(ctx context.Context, id string) (*FulfillmentModel.Outbound, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.Outbound
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OutboundRepository) GetByTaskID(ctx context.Context, taskID string) (*FulfillmentModel.Outbound, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.Outbound
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND task_id = ?", tenantUUID, strings.TrimSpace(taskID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OutboundRepository) Save(ctx context.Context, row *FulfillmentModel.Outbound) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *OutboundRepository) List(ctx context.Context, status string) ([]FulfillmentModel.Outbound, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []FulfillmentModel.Outbound
	err = db.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

type PickItemRepository struct {
	*Repository[FulfillmentModel.PickItem]
}

func NewPickItemRepository(db *gorm.DB) *PickItemRepository {
	return &PickItemRepository{Repository: NewRepository[FulfillmentModel.PickItem](db)}
}

func (r *PickItemRepository) CreateBatch(ctx context.Context, rows []FulfillmentModel.PickItem) error {
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
	return r.DB.WithContext(ctx).Create(&rows).Error
}

func (r *PickItemRepository) ListByOutboundID(ctx context.Context, outboundID string) ([]FulfillmentModel.PickItem, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []FulfillmentModel.PickItem
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND outbound_id = ?", tenantUUID, strings.TrimSpace(outboundID)).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

type PackOrderRepository struct {
	*Repository[FulfillmentModel.PackOrder]
}

func NewPackOrderRepository(db *gorm.DB) *PackOrderRepository {
	return &PackOrderRepository{Repository: NewRepository[FulfillmentModel.PackOrder](db)}
}

func (r *PackOrderRepository) Create(ctx context.Context, row *FulfillmentModel.PackOrder) error {
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

func (r *PackOrderRepository) GetByOutboundID(ctx context.Context, outboundID string) (*FulfillmentModel.PackOrder, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.PackOrder
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND outbound_id = ?", tenantUUID, strings.TrimSpace(outboundID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
