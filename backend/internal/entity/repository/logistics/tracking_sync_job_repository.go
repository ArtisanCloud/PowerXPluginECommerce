package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type TrackingSyncJobFilter struct {
	CarrierID string
	Status    string
	Limit     int
	Since     *time.Time
}

// TrackingSyncJobRepository stores tracking sync job states and metrics.
type TrackingSyncJobRepository struct {
	*Repository[LogisticsModel.TrackingSyncJob]
}

func NewTrackingSyncJobRepository(db *gorm.DB) *TrackingSyncJobRepository {
	return &TrackingSyncJobRepository{Repository: NewRepository[LogisticsModel.TrackingSyncJob](db)}
}

func (r *TrackingSyncJobRepository) Create(ctx context.Context, row *LogisticsModel.TrackingSyncJob) error {
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

func (r *TrackingSyncJobRepository) Save(ctx context.Context, row *LogisticsModel.TrackingSyncJob) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *TrackingSyncJobRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.TrackingSyncJob, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.TrackingSyncJob
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TrackingSyncJobRepository) List(ctx context.Context, filter TrackingSyncJobFilter) ([]LogisticsModel.TrackingSyncJob, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
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
	if filter.Since != nil {
		db = db.Where("created_at >= ?", filter.Since.UTC())
	}
	var rows []LogisticsModel.TrackingSyncJob
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
