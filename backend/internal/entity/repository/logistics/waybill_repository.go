package logistics

import (
	"context"
	"errors"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// WaybillRepository manages waybill persistence.
type WaybillRepository struct {
	*Repository[LogisticsModel.Waybill]
}

type WaybillSyncFilter struct {
	CarrierID string
	Status    string
	Limit     int
}

func NewWaybillRepository(db *gorm.DB) *WaybillRepository {
	return &WaybillRepository{Repository: NewRepository[LogisticsModel.Waybill](db)}
}

func (r *WaybillRepository) List(ctx context.Context) ([]LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *WaybillRepository) ListForSync(ctx context.Context, filter WaybillSyncFilter) ([]LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	db := r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		db = db.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.Waybill
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *WaybillRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaybillRepository) GetByWaybillNo(ctx context.Context, waybillNo string) (*LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_no = ?", tenantUUID, strings.TrimSpace(waybillNo)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaybillRepository) ListByOrderID(ctx context.Context, orderID string) ([]LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, strings.TrimSpace(orderID)).
		Order("package_no ASC, created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *WaybillRepository) GetByOrderAndPackageKey(ctx context.Context, orderID, packageKey string) (*LogisticsModel.Waybill, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.Waybill
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ? AND package_key = ?", tenantUUID, strings.TrimSpace(orderID), strings.TrimSpace(packageKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaybillRepository) NextPackageNo(ctx context.Context, orderID string) (int, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return 0, err
	}
	var maxNo int
	if err := r.DB.WithContext(ctx).
		Model(&LogisticsModel.Waybill{}).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, strings.TrimSpace(orderID)).
		Select("COALESCE(MAX(package_no), 0)").
		Scan(&maxNo).Error; err != nil {
		return 0, err
	}
	return maxNo + 1, nil
}

func (r *WaybillRepository) Create(ctx context.Context, waybill *LogisticsModel.Waybill) error {
	if waybill == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(waybill.TenantUUID) == "" {
		waybill.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(waybill).Error
}

func (r *WaybillRepository) Save(ctx context.Context, waybill *LogisticsModel.Waybill) error {
	if waybill == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(waybill).Error
}

// TrackingEventRepository manages tracking event persistence.
type TrackingEventRepository struct {
	*Repository[LogisticsModel.TrackingEvent]
}

func NewTrackingEventRepository(db *gorm.DB) *TrackingEventRepository {
	return &TrackingEventRepository{Repository: NewRepository[LogisticsModel.TrackingEvent](db)}
}

func (r *TrackingEventRepository) ListByWaybillID(ctx context.Context, waybillID string) ([]LogisticsModel.TrackingEvent, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.TrackingEvent
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *TrackingEventRepository) Create(ctx context.Context, event *LogisticsModel.TrackingEvent) error {
	if event == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(event.TenantUUID) == "" {
		event.TenantUUID = tenantUUID
	}
	if err := r.DB.WithContext(ctx).Create(event).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrackingEventRepository) ExistsByEventKey(ctx context.Context, waybillNo, eventID string) (bool, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.DB.WithContext(ctx).
		Model(&LogisticsModel.TrackingEvent{}).
		Where("tenant_uuid = ? AND waybill_no = ? AND event_id = ?", tenantUUID, strings.TrimSpace(waybillNo), strings.TrimSpace(eventID)).
		Count(&count).Error
	return count > 0, err
}

func IsDuplicateConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique")
}
