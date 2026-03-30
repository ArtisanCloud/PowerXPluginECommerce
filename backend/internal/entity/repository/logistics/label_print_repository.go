package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// LabelPrintRepository stores and queries label print tasks.
type LabelPrintRepository struct {
	*Repository[LogisticsModel.LabelPrintTask]
}

func NewLabelPrintRepository(db *gorm.DB) *LabelPrintRepository {
	return &LabelPrintRepository{Repository: NewRepository[LogisticsModel.LabelPrintTask](db)}
}

func (r *LabelPrintRepository) Create(ctx context.Context, task *LogisticsModel.LabelPrintTask) error {
	if task == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(task.TenantUUID) == "" {
		task.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(task).Error
}

func (r *LabelPrintRepository) Save(ctx context.Context, task *LogisticsModel.LabelPrintTask) error {
	if task == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(task).Error
}

func (r *LabelPrintRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.LabelPrintTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.LabelPrintTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LabelPrintRepository) List(ctx context.Context, status string, limit int) ([]LogisticsModel.LabelPrintTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	db := r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.LabelPrintTask
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *LabelPrintRepository) ListByIDs(ctx context.Context, ids []string) ([]LogisticsModel.LabelPrintTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	cleanIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			cleanIDs = append(cleanIDs, id)
		}
	}
	if len(cleanIDs) == 0 {
		return []LogisticsModel.LabelPrintTask{}, nil
	}
	var rows []LogisticsModel.LabelPrintTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id IN ?", tenantUUID, cleanIDs).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *LabelPrintRepository) RetryCandidates(ctx context.Context, ids []string, now time.Time) ([]LogisticsModel.LabelPrintTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("status = ?", "failed").
		Where("attempt_count < max_attempts")
	if len(ids) > 0 {
		cleanIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" {
				cleanIDs = append(cleanIDs, id)
			}
		}
		if len(cleanIDs) == 0 {
			return []LogisticsModel.LabelPrintTask{}, nil
		}
		db = db.Where("id IN ?", cleanIDs)
	} else {
		db = db.Where("retry_queued_at IS NULL OR retry_queued_at <= ?", now)
	}
	var rows []LogisticsModel.LabelPrintTask
	err = db.Order("created_at ASC").Find(&rows).Error
	return rows, err
}
