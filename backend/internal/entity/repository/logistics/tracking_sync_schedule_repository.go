package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type TrackingSyncScheduleRepository struct {
	*Repository[LogisticsModel.TrackingSyncSchedule]
}

func NewTrackingSyncScheduleRepository(db *gorm.DB) *TrackingSyncScheduleRepository {
	return &TrackingSyncScheduleRepository{Repository: NewRepository[LogisticsModel.TrackingSyncSchedule](db)}
}

func (r *TrackingSyncScheduleRepository) Create(ctx context.Context, row *LogisticsModel.TrackingSyncSchedule) error {
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

func (r *TrackingSyncScheduleRepository) Save(ctx context.Context, row *LogisticsModel.TrackingSyncSchedule) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *TrackingSyncScheduleRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.TrackingSyncSchedule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.TrackingSyncSchedule
	err = r.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TrackingSyncScheduleRepository) List(ctx context.Context, enabled *bool, limit int) ([]LogisticsModel.TrackingSyncSchedule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabled != nil {
		db = db.Where("enabled = ?", *enabled)
	}
	var rows []LogisticsModel.TrackingSyncSchedule
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *TrackingSyncScheduleRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]LogisticsModel.TrackingSyncSchedule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var rows []LogisticsModel.TrackingSyncSchedule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND enabled = ?", tenantUUID, true).
		Where("next_trigger_at IS NULL OR next_trigger_at <= ?", now.UTC()).
		Order("next_trigger_at ASC NULLS FIRST, created_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
