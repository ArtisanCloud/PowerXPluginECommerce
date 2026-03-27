package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type GatewayFailureEventFilter struct {
	CarrierID string
	WaybillNo string
	Status    string
	Limit     int
	Since     *time.Time
}

type GatewayFailureEventRepository struct {
	*Repository[LogisticsModel.GatewayFailureEvent]
}

func NewGatewayFailureEventRepository(db *gorm.DB) *GatewayFailureEventRepository {
	return &GatewayFailureEventRepository{Repository: NewRepository[LogisticsModel.GatewayFailureEvent](db)}
}

func (r *GatewayFailureEventRepository) Create(ctx context.Context, row *LogisticsModel.GatewayFailureEvent) error {
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

func (r *GatewayFailureEventRepository) Save(ctx context.Context, row *LogisticsModel.GatewayFailureEvent) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *GatewayFailureEventRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.GatewayFailureEvent, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.GatewayFailureEvent
	err = r.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GatewayFailureEventRepository) List(ctx context.Context, filter GatewayFailureEventFilter) ([]LogisticsModel.GatewayFailureEvent, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		db = db.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.WaybillNo) != "" {
		db = db.Where("waybill_no = ?", strings.TrimSpace(filter.WaybillNo))
	}
	if strings.TrimSpace(filter.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	if filter.Since != nil {
		db = db.Where("created_at >= ?", filter.Since.UTC())
	}
	var rows []LogisticsModel.GatewayFailureEvent
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
