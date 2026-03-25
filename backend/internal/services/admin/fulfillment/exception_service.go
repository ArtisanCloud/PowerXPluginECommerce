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

type ExceptionService struct {
	taskRepo      *FulfillmentRepo.TaskRepository
	logRepo       *FulfillmentRepo.TaskLogRepository
	exceptionRepo *FulfillmentRepo.ExceptionRepository
	emitter       *FulfillmentObs.Emitter
}

func NewExceptionService(deps *app.Deps) *ExceptionService {
	if deps == nil || deps.DB == nil {
		return &ExceptionService{}
	}
	return &ExceptionService{
		taskRepo:      FulfillmentRepo.NewTaskRepository(deps.DB),
		logRepo:       FulfillmentRepo.NewTaskLogRepository(deps.DB),
		exceptionRepo: FulfillmentRepo.NewExceptionRepository(deps.DB),
		emitter:       FulfillmentObs.NewEmitter(deps.RuntimeLogger(context.Background(), "fulfillment-exception", nil)),
	}
}

type ReportExceptionRequest struct {
	TaskID    string         `json:"task_id"`
	WaybillID string         `json:"waybill_id,omitempty"`
	Type      string         `json:"type"`
	Reason    string         `json:"reason,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func (s *ExceptionService) List(ctx context.Context, tenantUUID, status string) ([]FulfillmentModel.Exception, error) {
	if s == nil || s.exceptionRepo == nil {
		return nil, errors.New("exception service unavailable")
	}
	return s.exceptionRepo.List(withTenantContext(ctx, tenantUUID), status)
}

func (s *ExceptionService) Report(ctx context.Context, tenantUUID string, req ReportExceptionRequest) (*FulfillmentModel.Exception, error) {
	if s == nil || s.exceptionRepo == nil || s.taskRepo == nil {
		return nil, errors.New("exception service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.TaskID) == "" || strings.TrimSpace(req.Type) == "" {
		return nil, errors.New("task_id/type required")
	}
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, err
	}
	payload, _ := jsonBytes(req.Metadata, []byte("{}"))
	ex := &FulfillmentModel.Exception{
		ID:         utils.NewUUID(),
		TaskID:     strings.TrimSpace(req.TaskID),
		WaybillID:  strings.TrimSpace(req.WaybillID),
		Type:       strings.TrimSpace(strings.ToLower(req.Type)),
		Status:     "open",
		Reason:     strings.TrimSpace(req.Reason),
		Metadata:   datatypes.JSON(payload),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		TenantUUID: task.TenantUUID,
	}
	if err := s.exceptionRepo.Create(ctx, ex); err != nil {
		return nil, err
	}
	task.Status = "exception"
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	_ = s.logRepo.Create(ctx, &FulfillmentModel.TaskLog{
		ID:     utils.NewUUID(),
		TaskID: task.ID,
		Action: "exception.reported",
		Detail: datatypes.JSON(payload),
	})
	s.emit(task.TenantUUID, task.ID, ex.ID, ex.Status, "exception.report", "created")
	return ex, nil
}

func (s *ExceptionService) EscalateOverdue(ctx context.Context, tenantUUID string, now time.Time) ([]FulfillmentModel.Exception, error) {
	if s == nil || s.exceptionRepo == nil {
		return nil, errors.New("exception service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	cutoff := now.UTC().Add(-24 * time.Hour)
	rows, err := s.exceptionRepo.EscalateOpenBefore(ctx, cutoff, now.UTC())
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		_ = s.logRepo.Create(ctx, &FulfillmentModel.TaskLog{
			ID:     utils.NewUUID(),
			TaskID: row.TaskID,
			Action: "exception.escalated",
			Detail: datatypes.JSON([]byte(`{"reason":"no_first_action_within_24h"}`)),
		})
		s.emit(row.TenantUUID, row.TaskID, row.ID, row.Status, "exception.escalate", "updated")
	}
	return rows, nil
}

func (s *ExceptionService) emit(tenantUUID, taskID, exceptionID, status, action, result string) {
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
