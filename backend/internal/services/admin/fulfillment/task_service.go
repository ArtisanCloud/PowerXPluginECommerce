package fulfillment

import (
	"context"
	"errors"
	"strings"
	"time"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	FulfillmentRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/fulfillment"
	FulfillmentObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/fulfillment"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type TaskService struct {
	taskRepo *FulfillmentRepo.TaskRepository
	logRepo  *FulfillmentRepo.TaskLogRepository
	emitter  *FulfillmentObs.Emitter
}

func NewTaskService(deps *app.Deps) *TaskService {
	if deps == nil || deps.DB == nil {
		return &TaskService{}
	}
	return &TaskService{
		taskRepo: FulfillmentRepo.NewTaskRepository(deps.DB),
		logRepo:  FulfillmentRepo.NewTaskLogRepository(deps.DB),
		emitter:  FulfillmentObs.NewEmitter(deps.RuntimeLogger(context.Background(), "fulfillment-task", nil)),
	}
}

type CreateTaskRequest struct {
	OrderID     string         `json:"order_id"`
	WarehouseID string         `json:"warehouse_id"`
	AssignedTo  string         `json:"assigned_to,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type AdvanceTaskRequest struct {
	Status     string         `json:"status"`
	OperatorID string         `json:"operator_id,omitempty"`
	Detail     map[string]any `json:"detail,omitempty"`
}

func (s *TaskService) List(ctx context.Context, tenantUUID, status string) ([]FulfillmentModel.Task, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("task service unavailable")
	}
	return s.taskRepo.List(withTenantContext(ctx, tenantUUID), status)
}

func (s *TaskService) Create(ctx context.Context, tenantUUID string, req CreateTaskRequest) (*FulfillmentModel.Task, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("task service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.OrderID) == "" || strings.TrimSpace(req.WarehouseID) == "" {
		return nil, errors.New("order_id/warehouse_id required")
	}
	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	task := &FulfillmentModel.Task{
		ID:          utils.NewUUID(),
		OrderID:     strings.TrimSpace(req.OrderID),
		WarehouseID: strings.TrimSpace(req.WarehouseID),
		Status:      "pending",
		AssignedTo:  strings.TrimSpace(req.AssignedTo),
		Metadata:    datatypes.JSON(metadata),
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	_ = s.appendLog(ctx, task.ID, "task.created", req.AssignedTo, map[string]any{"status": task.Status})
	s.emit(task.TenantUUID, task.ID, "", task.Status, "task.create", "created")
	return task, nil
}

func (s *TaskService) Complete(ctx context.Context, tenantUUID, taskID string, req AdvanceTaskRequest) (*FulfillmentModel.Task, error) {
	return s.Advance(ctx, tenantUUID, taskID, AdvanceTaskRequest{
		Status:     "completed",
		OperatorID: req.OperatorID,
		Detail:     req.Detail,
	})
}

func (s *TaskService) Advance(ctx context.Context, tenantUUID, taskID string, req AdvanceTaskRequest) (*FulfillmentModel.Task, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("task service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	next := strings.TrimSpace(strings.ToLower(req.Status))
	if next == "" {
		return nil, errors.New("status required")
	}
	if !isTransitionAllowed(task.Status, next) {
		return nil, errors.New("invalid task status transition")
	}
	task.Status = next
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	_ = s.appendLog(ctx, task.ID, "task.status."+next, req.OperatorID, req.Detail)
	s.emit(task.TenantUUID, task.ID, "", task.Status, "task.advance", "updated")
	return task, nil
}

func (s *TaskService) appendLog(ctx context.Context, taskID, action, operatorID string, detail map[string]any) error {
	if s == nil || s.logRepo == nil {
		return nil
	}
	payload, _ := jsonBytes(detail, []byte("{}"))
	return s.logRepo.Create(ctx, &FulfillmentModel.TaskLog{
		ID:         utils.NewUUID(),
		TaskID:     strings.TrimSpace(taskID),
		Action:     strings.TrimSpace(action),
		OperatorID: strings.TrimSpace(operatorID),
		Detail:     datatypes.JSON(payload),
	})
}

func isTransitionAllowed(current, next string) bool {
	current = strings.TrimSpace(strings.ToLower(current))
	next = strings.TrimSpace(strings.ToLower(next))
	if current == "" || next == "" {
		return false
	}
	if current == next {
		return true
	}
	allowed := map[string][]string{
		"pending":   {"picking", "completed", "exception"},
		"picking":   {"packed", "completed", "exception"},
		"packed":    {"completed", "exception"},
		"exception": {"picking", "packed", "completed"},
	}
	for _, candidate := range allowed[current] {
		if candidate == next {
			return true
		}
	}
	return false
}

func (s *TaskService) emit(tenantUUID, taskID, exceptionID, status, action, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.Emit(FulfillmentObs.Event{
		Action:      strings.TrimSpace(action),
		TenantUUID:  strings.TrimSpace(tenantUUID),
		TaskID:      strings.TrimSpace(taskID),
		ExceptionID: strings.TrimSpace(exceptionID),
		Status:      strings.TrimSpace(status),
		Result:      strings.TrimSpace(result),
		EmittedAt:   time.Now().UTC(),
	})
}
