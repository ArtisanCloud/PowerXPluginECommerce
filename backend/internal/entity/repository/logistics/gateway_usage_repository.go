package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type GatewayUsageFilter struct {
	CarrierID string
	Provider  string
	From      *time.Time
	To        *time.Time
	Limit     int
}

type GatewayUsageRepository struct {
	*Repository[LogisticsModel.GatewayUsage]
}

func NewGatewayUsageRepository(db *gorm.DB) *GatewayUsageRepository {
	return &GatewayUsageRepository{Repository: NewRepository[LogisticsModel.GatewayUsage](db)}
}

func (r *GatewayUsageRepository) Create(ctx context.Context, row *LogisticsModel.GatewayUsage) error {
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

func (r *GatewayUsageRepository) GetBySourceJobID(ctx context.Context, sourceJobID string) (*LogisticsModel.GatewayUsage, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.GatewayUsage
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND source_job_id = ?", tenantUUID, strings.TrimSpace(sourceJobID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GatewayUsageRepository) List(ctx context.Context, filter GatewayUsageFilter) ([]LogisticsModel.GatewayUsage, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		db = db.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.Provider) != "" {
		db = db.Where("provider = ?", strings.TrimSpace(filter.Provider))
	}
	if filter.From != nil {
		db = db.Where("window_end_at >= ?", filter.From.UTC())
	}
	if filter.To != nil {
		db = db.Where("window_start_at <= ?", filter.To.UTC())
	}
	var rows []LogisticsModel.GatewayUsage
	err = db.Order("window_end_at DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
