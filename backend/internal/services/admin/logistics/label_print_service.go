package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

type LabelPrintService struct {
	taskRepo    *LogisticsRepo.LabelPrintRepository
	waybillRepo *LogisticsRepo.WaybillRepository
}

func NewLabelPrintService(deps *app.Deps) *LabelPrintService {
	if deps == nil || deps.DB == nil {
		return &LabelPrintService{}
	}
	return &LabelPrintService{
		taskRepo:    LogisticsRepo.NewLabelPrintRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
	}
}

type BatchPrintLabelsRequest struct {
	WaybillIDs     []string `json:"waybill_ids"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
	ReprintReason  string   `json:"reprint_reason,omitempty"`
	MaxAttempts    int      `json:"max_attempts,omitempty"`
}

type RetryLabelPrintRequest struct {
	TaskIDs []string `json:"task_ids,omitempty"`
}

type LabelPrintResult struct {
	Task              *LogisticsModel.LabelPrintTask `json:"task,omitempty"`
	IdempotencyStatus string                         `json:"idempotency_status"`
	Success           bool                           `json:"success"`
	Message           string                         `json:"message,omitempty"`
}

type BatchPrintLabelsResponse struct {
	Results []LabelPrintResult `json:"results"`
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
}

func (s *LabelPrintService) List(ctx context.Context, tenantUUID, status string, limit int) ([]LogisticsModel.LabelPrintTask, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("label print service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	return s.taskRepo.List(ctx, status, limit)
}

func (s *LabelPrintService) BatchPrint(ctx context.Context, tenantUUID string, req BatchPrintLabelsRequest) (*BatchPrintLabelsResponse, error) {
	if s == nil || s.taskRepo == nil || s.waybillRepo == nil {
		return nil, errors.New("label print service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	waybillIDs := normalizeStringList(req.WaybillIDs)
	if len(waybillIDs) == 0 {
		return nil, errors.New("waybill_ids required")
	}
	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	results := make([]LabelPrintResult, 0, len(waybillIDs))
	success, failed := 0, 0
	now := time.Now().UTC()
	for _, waybillID := range waybillIDs {
		wb, err := s.waybillRepo.GetByID(ctx, waybillID)
		if err != nil {
			results = append(results, LabelPrintResult{
				IdempotencyStatus: "created",
				Success:           false,
				Message:           "waybill not found",
			})
			failed++
			continue
		}
		requestKey := buildLabelRequestKey(strings.TrimSpace(req.IdempotencyKey), wb.ID)
		if replayed, err := s.taskRepo.GetByRequestKey(ctx, requestKey); err == nil && replayed != nil {
			results = append(results, LabelPrintResult{
				Task:              replayed,
				IdempotencyStatus: "replayed",
				Success:           strings.EqualFold(replayed.Status, "success"),
				Message:           "replayed by idempotency key",
			})
			if strings.EqualFold(replayed.Status, "success") {
				success++
			} else {
				failed++
			}
			continue
		}

		task := &LogisticsModel.LabelPrintTask{
			ID:          utils.NewUUID(),
			RequestKey:  requestKey,
			WaybillID:   wb.ID,
			WaybillNo:   wb.WaybillNo,
			Status:      "pending",
			MaxAttempts: maxAttempts,
			Metadata:    []byte("{}"),
		}
		printErr := s.tryPrintWaybill(wb)
		task.AttemptCount = 1
		if printErr != nil {
			task.Status = "failed"
			task.LastError = printErr.Error()
			retryAt := now.Add(2 * time.Minute)
			task.RetryQueuedAt = &retryAt
		} else {
			task.Status = "success"
			task.PrintedAt = &now
		}
		if strings.TrimSpace(req.ReprintReason) != "" {
			meta, _ := jsonBytes(map[string]any{
				"reprint_reason": strings.TrimSpace(req.ReprintReason),
				"source":         "batch",
			}, []byte("{}"))
			task.Metadata = meta
		}
		if err := s.taskRepo.Create(ctx, task); err != nil {
			return nil, err
		}
		ok := task.Status == "success"
		if ok {
			success++
		} else {
			failed++
		}
		results = append(results, LabelPrintResult{
			Task:              task,
			IdempotencyStatus: "created",
			Success:           ok,
			Message:           task.LastError,
		})
	}
	return &BatchPrintLabelsResponse{Results: results, Success: success, Failed: failed}, nil
}

func (s *LabelPrintService) RetryFailed(ctx context.Context, tenantUUID string, req RetryLabelPrintRequest) (*BatchPrintLabelsResponse, error) {
	if s == nil || s.taskRepo == nil || s.waybillRepo == nil {
		return nil, errors.New("label print service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	now := time.Now().UTC()
	candidates, err := s.taskRepo.RetryCandidates(ctx, normalizeStringList(req.TaskIDs), now)
	if err != nil {
		return nil, err
	}
	results := make([]LabelPrintResult, 0, len(candidates))
	success, failed := 0, 0
	for i := range candidates {
		task := &candidates[i]
		wb, err := s.waybillRepo.GetByID(ctx, task.WaybillID)
		if err != nil {
			task.AttemptCount++
			task.LastError = "waybill not found"
			queueAt := now.Add(5 * time.Minute)
			task.RetryQueuedAt = &queueAt
			_ = s.taskRepo.Save(ctx, task)
			failed++
			results = append(results, LabelPrintResult{
				Task:              task,
				IdempotencyStatus: "created",
				Success:           false,
				Message:           task.LastError,
			})
			continue
		}
		printErr := s.tryPrintWaybill(wb)
		task.AttemptCount++
		if printErr != nil {
			task.Status = "failed"
			task.LastError = printErr.Error()
			queueAt := now.Add(5 * time.Minute)
			task.RetryQueuedAt = &queueAt
			failed++
		} else {
			task.Status = "success"
			task.LastError = ""
			task.RetryQueuedAt = nil
			printedAt := now
			task.PrintedAt = &printedAt
			success++
		}
		if err := s.taskRepo.Save(ctx, task); err != nil {
			return nil, err
		}
		results = append(results, LabelPrintResult{
			Task:              task,
			IdempotencyStatus: "created",
			Success:           task.Status == "success",
			Message:           task.LastError,
		})
	}
	return &BatchPrintLabelsResponse{Results: results, Success: success, Failed: failed}, nil
}

func (s *LabelPrintService) tryPrintWaybill(wb *LogisticsModel.Waybill) error {
	if wb == nil {
		return errors.New("waybill is required")
	}
	if strings.TrimSpace(wb.LabelURL) == "" {
		return fmt.Errorf("waybill %s has empty label_url", strings.TrimSpace(wb.WaybillNo))
	}
	return nil
}

func normalizeStringList(raw []string) []string {
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func buildLabelRequestKey(idempotencyKey, waybillID string) string {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = "auto"
	}
	return fmt.Sprintf("%s#%s", idempotencyKey, strings.TrimSpace(waybillID))
}
