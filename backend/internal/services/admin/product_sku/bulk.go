package product_sku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SubmitBulkTask enqueues a price/inventory adjustment payload and triggers execution
// when审批不需要等待人工确认。
func (s *Service) SubmitBulkTask(ctx context.Context, req BulkAdjustmentRequest) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Operation.Type) == "" {
		return nil, errors.New("operation.type is required")
	}
	if !isSupportedOperation(req.Operation.Type) {
		return nil, fmt.Errorf("operation.type %q is not supported", req.Operation.Type)
	}
	if len(req.Operation.Value) == 0 {
		return nil, errors.New("operation.value is required")
	}
	scopeCount := len(req.Scope.SKUIDs)
	if scopeCount == 0 && len(req.Scope.Filters) == 0 {
		return nil, errors.New("scope.sku_ids or scope.filters is required")
	}

	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.TenantUUID != "" && req.TenantUUID != tenantID {
		return nil, fmt.Errorf("tenant mismatch: %s", req.TenantUUID)
	}
	if req.RequestedBy == "" {
		req.RequestedBy = "system"
	}

	approvalRequired, approvalReason := shouldRequireApproval(req)
	scopeJSON, err := json.Marshal(req.Scope)
	if err != nil {
		return nil, err
	}
	opJSON, err := json.Marshal(req.Operation)
	if err != nil {
		return nil, err
	}

	task := &productskumodel.ProductSKUBulkTask{
		TaskID:            utils.NewUUID(),
		TenantUUID:        tenantID,
		TaskType:          req.Operation.Type,
		Scope:             datatypes.JSON(scopeJSON),
		Operation:         datatypes.JSON(opJSON),
		ApprovalRequired:  approvalRequired,
		ApprovalState:     BulkTaskStatusPending,
		ApprovalReason:    approvalReason,
		Status:            BulkTaskStatusPending,
		AffectedCount:     scopeCount,
		SubmittedBy:       req.RequestedBy,
		ApprovalThreshold: approvalThreshold(req),
	}
	if !approvalRequired {
		task.ApprovalState = BulkTaskStatusApproved
		task.Status = BulkTaskStatusRunning
	}

	if err := s.persistBulkTask(ctx, tenantID, task, req.Scope.SKUIDs); err != nil {
		return nil, err
	}
	s.emitEvent(ctx, "bulk_task_submitted", map[string]any{
		"tenant_id":         tenantID,
		"task_id":           task.TaskID,
		"task_type":         task.TaskType,
		"approval_required": approvalRequired,
		"scope_size":        scopeCount,
	})

	var stats *BulkTaskStats
	if !approvalRequired {
		stats, err = s.executeBulkAdjustment(ctx, tenantID, task, req)
		if err != nil {
			return nil, err
		}
		s.emitEvent(ctx, "bulk_task_execution_completed", map[string]any{
			"tenant_id":  tenantID,
			"task_id":    task.TaskID,
			"task_type":  task.TaskType,
			"status":     task.Status,
			"succeeded":  stats.Succeeded,
			"failed":     stats.Failed,
			"auto_start": true,
		})
	}

	if stats != nil && !approvalRequired {
		_ = s.mergeTaskResultExtras(ctx, tenantID, task, map[string]any{
			"summary": fmt.Sprintf("%d sku(s) queued", stats.Succeeded),
		})
	}
	return s.buildBulkTaskResponse(ctx, tenantID, task)
}

type bulkTaskResultEnvelope struct {
	Stats  *BulkTaskStats `json:"stats,omitempty"`
	Result map[string]any `json:"result,omitempty"`
}

func encodeTaskResult(stats *BulkTaskStats, result map[string]any) datatypes.JSON {
	if stats == nil && len(result) == 0 {
		return nil
	}
	payload := bulkTaskResultEnvelope{
		Stats:  stats,
		Result: result,
	}
	body, _ := json.Marshal(payload)
	return datatypes.JSON(body)
}

func decodeTaskResult(data datatypes.JSON) (*BulkTaskStats, map[string]any) {
	if len(data) == 0 {
		return nil, nil
	}
	var envelope bulkTaskResultEnvelope
	if err := json.Unmarshal(data, &envelope); err == nil {
		return envelope.Stats, envelope.Result
	}
	var stats BulkTaskStats
	if err := json.Unmarshal(data, &stats); err == nil {
		return &stats, nil
	}
	return nil, nil
}

// GetBulkTask returns persisted task state and counters.
func (s *Service) GetBulkTask(ctx context.Context, taskID string) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task id is required")
	}

	var task productskumodel.ProductSKUBulkTask
	if err := s.BulkTaskRepo.DB.WithContext(ctx).
		Where("task_id = ? AND tenant_uuid = ?", taskID, tenantID).
		First(&task).Error; err != nil {
		return nil, err
	}

	return s.buildBulkTaskResponse(ctx, tenantID, &task)
}

// DecideBulkTaskApproval records reviewer decision and optionally triggers execution.
func (s *Service) DecideBulkTaskApproval(ctx context.Context, taskID string, decision BulkTaskApprovalDecision, actor string) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("task id is required")
	}
	choice := normalizeDecision(decision.Decision)
	if choice == "" {
		return nil, errors.New("decision must be approve or reject")
	}
	note := strings.TrimSpace(decision.Note)
	if choice == "reject" && note == "" {
		return nil, errors.New("note is required when rejecting a task")
	}
	var task productskumodel.ProductSKUBulkTask
	if err := s.BulkTaskRepo.DB.WithContext(ctx).
		Where("task_id = ? AND tenant_uuid = ?", taskID, tenantID).
		First(&task).Error; err != nil {
		return nil, err
	}
	if !task.ApprovalRequired {
		return nil, errors.New("task does not require approval")
	}
	if task.ApprovalState != BulkTaskStatusPending {
		return nil, fmt.Errorf("task already %s", task.ApprovalState)
	}
	if actor = strings.TrimSpace(actor); actor == "" {
		actor = "system"
	}
	now := time.Now()
	if choice == "reject" {
		task.ApprovalState = BulkTaskStatusRejected
		task.Status = BulkTaskStatusCancelled
		task.ApprovedBy = actor
		task.ApprovedAt = &now
		task.Result = encodeTaskResult(nil, map[string]any{
			"approval_note":     note,
			"approval_decision": choice,
		})
		if err := s.BulkTaskRepo.DB.WithContext(ctx).
			Model(&task).
			Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
			Updates(map[string]any{
				"approval_state": task.ApprovalState,
				"status":         task.Status,
				"approved_by":    task.ApprovedBy,
				"approved_at":    task.ApprovedAt,
				"result":         task.Result,
			}).Error; err != nil {
			return nil, err
		}
		s.emitEvent(ctx, "bulk_task_approval_decided", map[string]any{
			"tenant_id":         tenantID,
			"task_id":           task.TaskID,
			"decision":          choice,
			"actor":             actor,
			"approval_note":     note,
			"task_type":         task.TaskType,
			"status":            task.Status,
			"approval_required": task.ApprovalRequired,
		})
		return s.buildBulkTaskResponse(ctx, tenantID, &task)
	}

	task.ApprovalState = BulkTaskStatusApproved
	task.Status = BulkTaskStatusRunning
	task.ApprovedBy = actor
	task.ApprovedAt = &now

	if err := s.BulkTaskRepo.DB.WithContext(ctx).
		Model(&task).
		Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
		Updates(map[string]any{
			"approval_state": task.ApprovalState,
			"status":         task.Status,
			"approved_by":    task.ApprovedBy,
			"approved_at":    task.ApprovedAt,
		}).Error; err != nil {
		return nil, err
	}

	req, err := decodeBulkTaskRequest(task)
	if err != nil {
		return nil, err
	}
	stats, err := s.executeBulkAdjustment(ctx, tenantID, &task, req)
	if err != nil {
		return nil, err
	}
	if note != "" {
		_ = s.mergeTaskResultExtras(ctx, tenantID, &task, map[string]any{
			"approval_note":     note,
			"approval_decision": choice,
		})
	}
	s.emitEvent(ctx, "bulk_task_approval_decided", map[string]any{
		"tenant_id":         tenantID,
		"task_id":           task.TaskID,
		"decision":          choice,
		"actor":             actor,
		"approval_note":     note,
		"task_type":         task.TaskType,
		"status":            task.Status,
		"approval_required": task.ApprovalRequired,
		"succeeded":         statsSuccess(stats),
		"failed":            statsFailed(stats),
	})
	return s.buildBulkTaskResponse(ctx, tenantID, &task)
}

// RetryBulkTask re-runs a failed or completed bulk task.
func (s *Service) RetryBulkTask(ctx context.Context, taskID string) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("task id is required")
	}
	var task productskumodel.ProductSKUBulkTask
	if err := s.BulkTaskRepo.DB.WithContext(ctx).
		Where("task_id = ? AND tenant_uuid = ?", taskID, tenantID).
		First(&task).Error; err != nil {
		return nil, err
	}
	if task.ApprovalRequired && task.ApprovalState != BulkTaskStatusApproved {
		return nil, errors.New("task is pending approval")
	}
	if task.Status == BulkTaskStatusRunning {
		return nil, errors.New("task already running")
	}
	req, err := decodeBulkTaskRequest(task)
	if err != nil {
		return nil, err
	}
	if err := s.resetBulkTaskItems(ctx, tenantID, task.TaskID); err != nil {
		return nil, err
	}
	task.Status = BulkTaskStatusRunning
	task.Result = nil
	task.ErrorReportURL = ""
	if err := s.BulkTaskRepo.DB.WithContext(ctx).
		Model(&task).
		Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
		Updates(map[string]any{
			"status":           task.Status,
			"result":           task.Result,
			"error_report_url": task.ErrorReportURL,
		}).Error; err != nil {
		return nil, err
	}
	stats, err := s.executeBulkAdjustment(ctx, tenantID, &task, req)
	if err != nil {
		return nil, err
	}
	s.emitEvent(ctx, "bulk_task_retried", map[string]any{
		"tenant_id": tenantID,
		"task_id":   task.TaskID,
		"task_type": task.TaskType,
		"status":    task.Status,
		"succeeded": statsSuccess(stats),
		"failed":    statsFailed(stats),
	})
	return s.buildBulkTaskResponse(ctx, tenantID, &task)
}

func (s *Service) persistBulkTask(ctx context.Context, tenantID string, task *productskumodel.ProductSKUBulkTask, skuIDs []string) error {
	return s.BulkTaskRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		if len(skuIDs) == 0 {
			return nil
		}
		items := make([]productskumodel.ProductSKUBulkTaskItem, 0, len(skuIDs))
		for _, id := range skuIDs {
			cleanID := strings.TrimSpace(id)
			if cleanID == "" {
				continue
			}
			items = append(items, productskumodel.ProductSKUBulkTaskItem{
				TenantUUID: tenantID,
				TaskID:     task.TaskID,
				SKUId:      cleanID,
				Status:     BulkTaskStatusPending,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func statsSuccess(stats *BulkTaskStats) int {
	if stats == nil {
		return 0
	}
	return stats.Succeeded
}

func statsFailed(stats *BulkTaskStats) int {
	if stats == nil {
		return 0
	}
	return stats.Failed
}

func (s *Service) executeBulkAdjustment(ctx context.Context, tenantID string, task *productskumodel.ProductSKUBulkTask, req BulkAdjustmentRequest) (*BulkTaskStats, error) {
	stats := &BulkTaskStats{}
	err := s.BulkTaskRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		if err := tx.Model(task).
			Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
			Update("status", BulkTaskStatusRunning).Error; err != nil {
			return err
		}

		var diff datatypes.JSON
		if len(req.Operation.Value) > 0 {
			diff = datatypes.JSON(req.Operation.Value)
		}
		if err := tx.Model(&productskumodel.ProductSKUBulkTaskItem{}).
			Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
			Updates(map[string]any{
				"status": BulkTaskStatusSucceeded,
				"diff":   diff,
			}).Error; err != nil {
			return err
		}

		stats.Succeeded = len(req.Scope.SKUIDs)
		task.Status = BulkTaskStatusSucceeded
		task.Result = encodeTaskResult(stats, nil)
		task.AffectedCount = stats.Succeeded
		return tx.Model(task).
			Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
			Updates(map[string]any{
				"status":         task.Status,
				"result":         task.Result,
				"affected_count": task.AffectedCount,
			}).Error
	})
	if err != nil {
		_ = s.BulkTaskRepo.DB.WithContext(ctx).
			Model(task).
			Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
			Update("status", BulkTaskStatusFailed)
		return nil, err
	}
	return stats, nil
}

func (s *Service) loadBulkTaskStats(ctx context.Context, tenantID string, task productskumodel.ProductSKUBulkTask) (*BulkTaskStats, error) {
	if stats, _ := decodeTaskResult(task.Result); stats != nil {
		return stats, nil
	}
	var rows []struct {
		Status string
		Count  int64
	}
	if err := s.BulkTaskItemRepo.DB.WithContext(ctx).
		Model(&productskumodel.ProductSKUBulkTaskItem{}).
		Select("status, COUNT(*) as count").
		Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	stats := &BulkTaskStats{}
	for _, row := range rows {
		switch strings.ToLower(row.Status) {
		case BulkTaskStatusSucceeded:
			stats.Succeeded += int(row.Count)
		case BulkTaskStatusFailed:
			stats.Failed += int(row.Count)
		}
	}
	return stats, nil
}

func (s *Service) buildBulkTaskResponse(ctx context.Context, tenantID string, task *productskumodel.ProductSKUBulkTask) (*BulkTaskResponse, error) {
	if task == nil {
		return nil, errors.New("task is required")
	}
	var scope map[string]any
	if len(task.Scope) > 0 {
		_ = json.Unmarshal(task.Scope, &scope)
	}
	var operation map[string]any
	if len(task.Operation) > 0 {
		_ = json.Unmarshal(task.Operation, &operation)
	}
	stats, result := decodeTaskResult(task.Result)
	if stats == nil {
		var err error
		stats, err = s.loadBulkTaskStats(ctx, tenantID, *task)
		if err != nil {
			return nil, err
		}
	}
	resp := &BulkTaskResponse{
		TaskID:            task.TaskID,
		TaskType:          task.TaskType,
		Scope:             scope,
		Operation:         operation,
		ApprovalRequired:  task.ApprovalRequired,
		ApprovalState:     task.ApprovalState,
		ApprovalReason:    task.ApprovalReason,
		ApprovalThreshold: task.ApprovalThreshold,
		Status:            task.Status,
		AffectedCount:     task.AffectedCount,
		SubmittedBy:       task.SubmittedBy,
		ApprovedBy:        task.ApprovedBy,
		ErrorReportURL:    task.ErrorReportURL,
		Stats:             stats,
		Result:            result,
	}
	if !task.CreatedAt.IsZero() {
		createdAt := task.CreatedAt
		resp.CreatedAt = &createdAt
	}
	if !task.UpdatedAt.IsZero() {
		updatedAt := task.UpdatedAt
		resp.UpdatedAt = &updatedAt
	}
	if task.ApprovedAt != nil && !task.ApprovedAt.IsZero() {
		resp.ApprovedAt = task.ApprovedAt
	}
	return resp, nil
}

func decodeBulkTaskRequest(task productskumodel.ProductSKUBulkTask) (BulkAdjustmentRequest, error) {
	req := BulkAdjustmentRequest{}
	if len(task.Scope) > 0 {
		if err := json.Unmarshal(task.Scope, &req.Scope); err != nil {
			return req, err
		}
	}
	req.Operation.Type = task.TaskType
	if len(task.Operation) > 0 {
		req.Operation.Value = append(json.RawMessage(nil), json.RawMessage(task.Operation)...)
	}
	return req, nil
}

func (s *Service) resetBulkTaskItems(ctx context.Context, tenantID, taskID string) error {
	return s.BulkTaskItemRepo.DB.WithContext(ctx).
		Model(&productskumodel.ProductSKUBulkTaskItem{}).
		Where("task_id = ? AND tenant_uuid = ?", taskID, tenantID).
		Updates(map[string]any{
			"status":  BulkTaskStatusPending,
			"message": "",
			"diff":    datatypes.JSON([]byte("null")),
		}).Error
}

func (s *Service) mergeTaskResultExtras(ctx context.Context, tenantID string, task *productskumodel.ProductSKUBulkTask, extra map[string]any) error {
	if task == nil || len(extra) == 0 {
		return nil
	}
	stats, result := decodeTaskResult(task.Result)
	if result == nil {
		result = make(map[string]any)
	}
	for k, v := range extra {
		result[k] = v
	}
	task.Result = encodeTaskResult(stats, result)
	return s.BulkTaskRepo.DB.WithContext(ctx).
		Model(task).
		Where("task_id = ? AND tenant_uuid = ?", task.TaskID, tenantID).
		Update("result", task.Result).Error
}

func normalizeDecision(decision string) string {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "approve", "approved":
		return "approve"
	case "reject", "rejected":
		return "reject"
	default:
		return ""
	}
}

func isSupportedOperation(op string) bool {
	switch op {
	case BulkOpPriceFixed, BulkOpPricePercent, BulkOpInventoryFixed, BulkOpInventoryReplace:
		return true
	default:
		return false
	}
}

func shouldRequireApproval(req BulkAdjustmentRequest) (bool, string) {
	reasons := make([]string, 0, 3)
	if req.ApprovalContext != nil {
		if req.ApprovalContext.ThresholdAmount > 0 {
			reasons = append(reasons, "threshold_amount")
		}
		if strings.TrimSpace(req.ApprovalContext.Reason) != "" {
			reasons = append(reasons, "reason_provided")
		}
	}
	if len(req.Scope.SKUIDs) >= 50 {
		reasons = append(reasons, "bulk_size>=50")
	}
	if req.Operation.Type == BulkOpPricePercent {
		if delta := math.Abs(parseNumeric(req.Operation.Value)); delta >= 5 {
			reasons = append(reasons, "price_percent>=5")
		}
	}
	return len(reasons) > 0, strings.Join(reasons, ",")
}

func approvalThreshold(req BulkAdjustmentRequest) float64 {
	if req.ApprovalContext != nil && req.ApprovalContext.ThresholdAmount > 0 {
		return req.ApprovalContext.ThresholdAmount
	}
	return 0
}

func parseNumeric(raw json.RawMessage) float64 {
	if len(raw) == 0 {
		return 0
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	return 0
}
