package reverse

import (
	"context"
	"strings"

	ReverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	"gorm.io/gorm"
)

// WaybillRepository manages reverse-waybill persistence.
type WaybillRepository struct {
	*Repository[ReverseModel.Waybill]
}

func NewWaybillRepository(db *gorm.DB) *WaybillRepository {
	return &WaybillRepository{Repository: NewRepository[ReverseModel.Waybill](db)}
}

func (r *WaybillRepository) List(ctx context.Context) ([]ReverseModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ReverseModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *WaybillRepository) GetByID(ctx context.Context, id string) (*ReverseModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row ReverseModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaybillRepository) Create(ctx context.Context, row *ReverseModel.Waybill) error {
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

func (r *WaybillRepository) Save(ctx context.Context, row *ReverseModel.Waybill) error {
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

// TrackingEventRepository manages reverse-waybill tracking events.
type TrackingEventRepository struct {
	*Repository[ReverseModel.TrackingEvent]
}

func NewTrackingEventRepository(db *gorm.DB) *TrackingEventRepository {
	return &TrackingEventRepository{Repository: NewRepository[ReverseModel.TrackingEvent](db)}
}

func (r *TrackingEventRepository) ListByWaybillID(ctx context.Context, waybillID string) ([]ReverseModel.TrackingEvent, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ReverseModel.TrackingEvent
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *TrackingEventRepository) Create(ctx context.Context, row *ReverseModel.TrackingEvent) error {
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

// WarehouseResultRepository manages inbound-inspection records.
type WarehouseResultRepository struct {
	*Repository[ReverseModel.WarehouseResult]
}

func NewWarehouseResultRepository(db *gorm.DB) *WarehouseResultRepository {
	return &WarehouseResultRepository{Repository: NewRepository[ReverseModel.WarehouseResult](db)}
}

func (r *WarehouseResultRepository) ListByWaybillID(ctx context.Context, waybillID string) ([]ReverseModel.WarehouseResult, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ReverseModel.WarehouseResult
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *WarehouseResultRepository) Create(ctx context.Context, row *ReverseModel.WarehouseResult) error {
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
