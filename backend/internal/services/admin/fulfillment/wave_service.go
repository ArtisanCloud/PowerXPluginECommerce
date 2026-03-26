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

type WaveService struct {
	waveRepo *FulfillmentRepo.WaveRepository
	linkRepo *FulfillmentRepo.WaveTaskLinkRepository
	taskRepo *FulfillmentRepo.TaskRepository
	logRepo  *FulfillmentRepo.TaskLogRepository
	emitter  *FulfillmentObs.Emitter
}

func NewWaveService(deps *app.Deps) *WaveService {
	if deps == nil || deps.DB == nil {
		return &WaveService{}
	}
	return &WaveService{
		waveRepo: FulfillmentRepo.NewWaveRepository(deps.DB),
		linkRepo: FulfillmentRepo.NewWaveTaskLinkRepository(deps.DB),
		taskRepo: FulfillmentRepo.NewTaskRepository(deps.DB),
		logRepo:  FulfillmentRepo.NewTaskLogRepository(deps.DB),
		emitter:  FulfillmentObs.NewEmitter(deps.RuntimeLogger(context.Background(), "fulfillment-wave", nil)),
	}
}

type CreateWaveRequest struct {
	Name        string   `json:"name"`
	WarehouseID string   `json:"warehouse_id"`
	TaskIDs     []string `json:"task_ids"`
	Metadata    map[string]any
}

type BatchAdvanceWaveRequest struct {
	Status      string   `json:"status"`
	OperatorID  string   `json:"operator_id,omitempty"`
	FailTaskIDs []string `json:"fail_task_ids,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

type ReassignWaveTaskRequest struct {
	TaskID     string `json:"task_id"`
	AssignedTo string `json:"assigned_to"`
	OperatorID string `json:"operator_id,omitempty"`
}

type WaveBatchResult struct {
	WaveID      string                          `json:"wave_id"`
	WaveStatus  string                          `json:"wave_status"`
	Success     int                             `json:"success"`
	Failed      int                             `json:"failed"`
	TaskResults []FulfillmentModel.WaveTaskLink `json:"task_results"`
}

type WaveDetail struct {
	Wave  *FulfillmentModel.Wave          `json:"wave"`
	Tasks []FulfillmentModel.WaveTaskLink `json:"tasks"`
}

func (s *WaveService) List(ctx context.Context, tenantUUID, status string) ([]FulfillmentModel.Wave, error) {
	if s == nil || s.waveRepo == nil {
		return nil, errors.New("wave service unavailable")
	}
	return s.waveRepo.List(withTenantContext(ctx, tenantUUID), status)
}

func (s *WaveService) Detail(ctx context.Context, tenantUUID, waveID string) (*WaveDetail, error) {
	if s == nil || s.waveRepo == nil || s.linkRepo == nil {
		return nil, errors.New("wave service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wave, err := s.waveRepo.GetByID(ctx, waveID)
	if err != nil {
		return nil, err
	}
	links, err := s.linkRepo.ListByWaveID(ctx, waveID)
	if err != nil {
		return nil, err
	}
	return &WaveDetail{Wave: wave, Tasks: links}, nil
}

func (s *WaveService) Create(ctx context.Context, tenantUUID string, req CreateWaveRequest) (*WaveDetail, error) {
	if s == nil || s.waveRepo == nil || s.linkRepo == nil || s.taskRepo == nil {
		return nil, errors.New("wave service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	taskIDs := normalizeTaskIDs(req.TaskIDs)
	if strings.TrimSpace(req.WarehouseID) == "" || len(taskIDs) == 0 {
		return nil, errors.New("warehouse_id/task_ids required")
	}
	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	wave := &FulfillmentModel.Wave{
		ID:          utils.NewUUID(),
		Name:        defaultWaveName(strings.TrimSpace(req.Name), strings.TrimSpace(req.WarehouseID)),
		WarehouseID: strings.TrimSpace(req.WarehouseID),
		Status:      "pending",
		Metadata:    datatypes.JSON(metadata),
	}
	if err := s.waveRepo.Create(ctx, wave); err != nil {
		return nil, err
	}
	links := make([]FulfillmentModel.WaveTaskLink, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, err := s.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			continue
		}
		if strings.TrimSpace(task.WarehouseID) != strings.TrimSpace(req.WarehouseID) {
			continue
		}
		link := FulfillmentModel.WaveTaskLink{
			ID:       utils.NewUUID(),
			WaveID:   wave.ID,
			TaskID:   taskID,
			Result:   "pending",
			Message:  "",
			Metadata: datatypes.JSON([]byte("{}")),
		}
		if err := s.linkRepo.Create(ctx, &link); err != nil {
			continue
		}
		links = append(links, link)
	}
	if len(links) == 0 {
		return nil, errors.New("no eligible tasks for wave")
	}
	return &WaveDetail{Wave: wave, Tasks: links}, nil
}

func (s *WaveService) BatchAdvance(ctx context.Context, tenantUUID, waveID string, req BatchAdvanceWaveRequest) (*WaveBatchResult, error) {
	if s == nil || s.waveRepo == nil || s.linkRepo == nil || s.taskRepo == nil {
		return nil, errors.New("wave service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	next := strings.TrimSpace(strings.ToLower(req.Status))
	if next == "" {
		return nil, errors.New("status required")
	}
	wave, err := s.waveRepo.GetByID(ctx, waveID)
	if err != nil {
		return nil, err
	}
	links, err := s.linkRepo.ListByWaveID(ctx, waveID)
	if err != nil {
		return nil, err
	}
	failSet := make(map[string]struct{}, len(req.FailTaskIDs))
	for _, id := range req.FailTaskIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			failSet[id] = struct{}{}
		}
	}
	successCount := 0
	failedCount := 0
	for i := range links {
		link := &links[i]
		task, err := s.taskRepo.GetByID(ctx, link.TaskID)
		if err != nil {
			failedCount++
			link.Result = "failed"
			link.Message = "task not found"
			_ = s.linkRepo.Save(ctx, link)
			continue
		}
		if _, shouldFail := failSet[task.ID]; shouldFail {
			failedCount++
			link.Result = "failed"
			link.Message = fallbackString(strings.TrimSpace(req.Reason), "forced failure")
			_ = s.linkRepo.Save(ctx, link)
			_ = s.appendTaskLog(ctx, task.ID, "task.wave.batch."+next, req.OperatorID, map[string]any{
				"wave_id": wave.ID,
				"result":  "failed",
				"reason":  link.Message,
			})
			continue
		}
		if !isTransitionAllowed(task.Status, next) {
			failedCount++
			link.Result = "failed"
			link.Message = "invalid task status transition"
			_ = s.linkRepo.Save(ctx, link)
			_ = s.appendTaskLog(ctx, task.ID, "task.wave.batch."+next, req.OperatorID, map[string]any{
				"wave_id": wave.ID,
				"result":  "failed",
				"reason":  link.Message,
			})
			continue
		}
		task.Status = next
		if err := s.taskRepo.Save(ctx, task); err != nil {
			failedCount++
			link.Result = "failed"
			link.Message = err.Error()
			_ = s.linkRepo.Save(ctx, link)
			continue
		}
		successCount++
		link.Result = "success"
		link.Message = ""
		_ = s.linkRepo.Save(ctx, link)
		_ = s.appendTaskLog(ctx, task.ID, "task.wave.batch."+next, req.OperatorID, map[string]any{
			"wave_id": wave.ID,
			"result":  "success",
			"status":  next,
		})
		s.emit(task.TenantUUID, task.ID, next, "wave.batch.advance", "updated")
	}
	wave.Status = resolveWaveStatus(next, successCount, failedCount)
	if err := s.waveRepo.Save(ctx, wave); err != nil {
		return nil, err
	}
	return &WaveBatchResult{
		WaveID:      wave.ID,
		WaveStatus:  wave.Status,
		Success:     successCount,
		Failed:      failedCount,
		TaskResults: links,
	}, nil
}

func (s *WaveService) ReassignTask(ctx context.Context, tenantUUID, waveID string, req ReassignWaveTaskRequest) (*FulfillmentModel.Task, error) {
	if s == nil || s.linkRepo == nil || s.taskRepo == nil {
		return nil, errors.New("wave service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	link, err := s.linkRepo.GetByWaveAndTask(ctx, waveID, req.TaskID)
	if err != nil {
		return nil, err
	}
	task, err := s.taskRepo.GetByID(ctx, link.TaskID)
	if err != nil {
		return nil, err
	}
	task.AssignedTo = strings.TrimSpace(req.AssignedTo)
	if task.AssignedTo == "" {
		return nil, errors.New("assigned_to required")
	}
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	link.Result = "reassigned"
	link.Message = "task reassigned"
	_ = s.linkRepo.Save(ctx, link)
	_ = s.appendTaskLog(ctx, task.ID, "task.wave.reassign", req.OperatorID, map[string]any{
		"wave_id":     waveID,
		"assigned_to": task.AssignedTo,
	})
	s.emit(task.TenantUUID, task.ID, task.Status, "wave.task.reassign", "updated")
	return task, nil
}

func (s *WaveService) appendTaskLog(ctx context.Context, taskID, action, operatorID string, detail map[string]any) error {
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

func (s *WaveService) emit(tenantUUID, taskID, status, action, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.Emit(FulfillmentObs.Event{
		Action:     strings.TrimSpace(action),
		TenantUUID: strings.TrimSpace(tenantUUID),
		TaskID:     strings.TrimSpace(taskID),
		Status:     strings.TrimSpace(status),
		Result:     strings.TrimSpace(result),
		EmittedAt:  time.Now().UTC(),
	})
}

func normalizeTaskIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		val := strings.TrimSpace(id)
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		out = append(out, val)
	}
	return out
}

func defaultWaveName(name, warehouseID string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	return "wave-" + strings.TrimSpace(warehouseID) + "-" + time.Now().UTC().Format("20060102150405")
}

func resolveWaveStatus(next string, successCount, failedCount int) string {
	if successCount > 0 && failedCount > 0 {
		return "partially_failed"
	}
	if successCount == 0 && failedCount > 0 {
		return "failed"
	}
	if next == "completed" {
		return "completed"
	}
	return "running"
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
