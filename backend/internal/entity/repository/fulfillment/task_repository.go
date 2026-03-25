package fulfillment

import (
	"context"
	"strings"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	"gorm.io/gorm"
)

// TaskRepository manages fulfillment task persistence.
type TaskRepository struct {
	*Repository[FulfillmentModel.Task]
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{Repository: NewRepository[FulfillmentModel.Task](db)}
}

func (r *TaskRepository) List(ctx context.Context, status string) ([]FulfillmentModel.Task, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []FulfillmentModel.Task
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*FulfillmentModel.Task, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.Task
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TaskRepository) Create(ctx context.Context, task *FulfillmentModel.Task) error {
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

func (r *TaskRepository) Save(ctx context.Context, task *FulfillmentModel.Task) error {
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
	return r.DB.WithContext(ctx).Save(task).Error
}

// TaskLogRepository manages task-operation logs.
type TaskLogRepository struct {
	*Repository[FulfillmentModel.TaskLog]
}

func NewTaskLogRepository(db *gorm.DB) *TaskLogRepository {
	return &TaskLogRepository{Repository: NewRepository[FulfillmentModel.TaskLog](db)}
}

func (r *TaskLogRepository) Create(ctx context.Context, log *FulfillmentModel.TaskLog) error {
	if log == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(log.TenantUUID) == "" {
		log.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(log).Error
}

func (r *TaskLogRepository) ListByTaskID(ctx context.Context, taskID string) ([]FulfillmentModel.TaskLog, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []FulfillmentModel.TaskLog
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND task_id = ?", tenantUUID, strings.TrimSpace(taskID)).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}
