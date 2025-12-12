package spu

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	taskcenter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	importJobType = "product_spu_import"
	exportJobType = "product_spu_export"
)

// ImportRequest wraps the payload received from HTTP upload.
type ImportRequest struct {
	TemplateID string
	Filename   string
	Payload    []byte
}

// ExportRequest captures filters and columns required for export.
type ExportRequest struct {
	Filters map[string]any `json:"filters"`
	Fields  []string       `json:"fields"`
}

// ImportService orchestrates bulk import/export flows and reports via job store.
type ImportService struct {
	deps     *app.Deps
	jobStore *taskcenter.Store
}

// NewImportService constructs the orchestrator with shared dependencies.
func NewImportService(deps *app.Deps) *ImportService {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &ImportService{
		deps:     deps,
		jobStore: taskcenter.DefaultStore(),
	}
}

// StartImport registers an import job and schedules asynchronous processing.
func (s *ImportService) StartImport(ctx context.Context, req ImportRequest) (string, error) {
	if s == nil {
		return "", errors.New("import service unavailable")
	}
	if len(req.Payload) == 0 {
		return "", errors.New("导入文件内容为空")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(req.TemplateID) == "" {
		req.TemplateID = "default"
	}
	job := s.jobStore.Create(importJobType, map[string]any{
		"tenantUuid": tenantID,
		"templateId": req.TemplateID,
		"filename":   req.Filename,
	})
	if job == nil {
		return "", errors.New("无法创建导入任务")
	}
	if err := s.insertImportTask(ctx, tenantID, job.TaskID, req); err != nil {
		return "", err
	}
	s.emitAudit(ctx, tenantID, job.TaskID, "import_start", map[string]any{
		"templateId": req.TemplateID,
		"filename":   req.Filename,
	})
	go s.runImportJob(authx.ContextWithTenantUUID(context.Background(), tenantID), job.TaskID, req)
	return job.TaskID, nil
}

// StartExport registers an export job and creates a CSV snapshot asynchronously.
func (s *ImportService) StartExport(ctx context.Context, req ExportRequest) (string, error) {
	if s == nil {
		return "", errors.New("import service unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return "", err
	}
	job := s.jobStore.Create(exportJobType, map[string]any{
		"tenantUuid": tenantID,
		"fields":     req.Fields,
	})
	if job == nil {
		return "", errors.New("无法创建导出任务")
	}
	if err := s.insertExportTask(ctx, tenantID, job.TaskID, req); err != nil {
		return "", err
	}
	s.emitAudit(ctx, tenantID, job.TaskID, "export_start", map[string]any{"fields": req.Fields})
	go s.runExportJob(authx.ContextWithTenantUUID(context.Background(), tenantID), job.TaskID, req)
	return job.TaskID, nil
}

func (s *ImportService) runImportJob(ctx context.Context, taskID string, req ImportRequest) {
	_, _ = s.jobStore.SetStatus(taskID, "running", "正在导入商品")
	_ = s.updateImportTaskStatus(ctx, taskID, "running", nil)
	records, err := parseImportRecords(req.Payload)
	if err != nil {
		s.failImport(ctx, taskID, err)
		return
	}
	summary := validateImportRecords(records)
	reportPath := ""
	if len(summary.Failures) > 0 {
		if path, err := persistFailureReport(taskID, summary.Failures); err == nil {
			reportPath = path
		}
	}
	if err := s.finalizeImportTask(ctx, taskID, summary, reportPath); err != nil {
		s.failImport(ctx, taskID, err)
		return
	}
	message := fmt.Sprintf("成功 %d 行，失败 %d 行", summary.SuccessCount, len(summary.Failures))
	_, _ = s.jobStore.Success(taskID, message, reportPath)
	tenantID, _ := authx.TenantUUIDFromContext(ctx)
	s.emitAudit(ctx, tenantID, taskID, "import_completed", map[string]any{
		"success": summary.SuccessCount,
		"failed":  len(summary.Failures),
		"report":  reportPath,
	})
}

func (s *ImportService) runExportJob(ctx context.Context, taskID string, req ExportRequest) {
	_, _ = s.jobStore.SetStatus(taskID, "running", "正在导出商品")
	_ = s.updateExportTaskStatus(ctx, taskID, "running", "")
	path, err := s.buildExportSnapshot(ctx, taskID, req)
	if err != nil {
		s.failExport(ctx, taskID, err)
		return
	}
	message := "导出完成，可下载文件"
	_, _ = s.jobStore.Success(taskID, message, path)
	_ = s.completeExportTask(ctx, taskID, "success", path)
	tenantID, _ := authx.TenantUUIDFromContext(ctx)
	s.emitAudit(ctx, tenantID, taskID, "export_completed", map[string]any{
		"downloadUrl": path,
	})
}

func (s *ImportService) failImport(ctx context.Context, taskID string, err error) {
	_, _ = s.jobStore.Fail(taskID, err)
	_ = s.updateImportTaskStatus(ctx, taskID, "failed", err)
	tenantID, _ := authx.TenantUUIDFromContext(ctx)
	s.emitAudit(ctx, tenantID, taskID, "import_failed", map[string]any{"error": err.Error()})
}

func (s *ImportService) failExport(ctx context.Context, taskID string, err error) {
	_, _ = s.jobStore.Fail(taskID, err)
	_ = s.completeExportTask(ctx, taskID, "failed", "")
	tenantID, _ := authx.TenantUUIDFromContext(ctx)
	s.emitAudit(ctx, tenantID, taskID, "export_failed", map[string]any{"error": err.Error()})
}

func (s *ImportService) insertImportTask(ctx context.Context, tenantID, taskID string, req ImportRequest) error {
	entry := productmodel.SPUImportTask{
		TaskID:     taskID,
		TenantUUID: tenantID,
		Template:   req.TemplateID,
		Status:     "queued",
		Initiator:  actorFromContext(ctx),
		Detail:     encodeMetadata(map[string]any{"filename": req.Filename}),
	}
	return s.deps.DB.WithContext(ctx).Create(&entry).Error
}

func (s *ImportService) insertExportTask(ctx context.Context, tenantID, taskID string, req ExportRequest) error {
	entry := productmodel.SPUExportTask{
		TaskID:     taskID,
		TenantUUID: tenantID,
		Status:     "queued",
		Initiator:  actorFromContext(ctx),
		Filters:    encodeMetadata(req.Filters),
		Fields:     pq.StringArray(req.Fields),
	}
	return s.deps.DB.WithContext(ctx).Create(&entry).Error
}

func (s *ImportService) updateImportTaskStatus(ctx context.Context, taskID, status string, failure error) error {
	updates := map[string]any{"status": status}
	if failure != nil {
		updates["failed_report"] = failure.Error()
	}
	return s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPUImportTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func (s *ImportService) finalizeImportTask(ctx context.Context, taskID string, summary ImportSummary, report string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":       statusFromSummary(summary),
		"success_rows": summary.SuccessCount,
		"failed_rows":  len(summary.Failures),
		"completed_at": &now,
		"failed_report": func() string {
			if report == "" {
				if len(summary.Failures) == 0 {
					return ""
				}
				body, _ := json.Marshal(summary.Failures)
				return string(body)
			}
			return report
		}(),
	}
	return s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPUImportTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func statusFromSummary(summary ImportSummary) string {
	if summary.SuccessCount == 0 && summary.TotalRows > 0 {
		return "failed"
	}
	if len(summary.Failures) > 0 {
		return "partial"
	}
	return "success"
}

func (s *ImportService) completeExportTask(ctx context.Context, taskID, status, download string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":       status,
		"download_url": download,
		"completed_at": &now,
	}
	return s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPUExportTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func (s *ImportService) buildExportSnapshot(ctx context.Context, taskID string, req ExportRequest) (string, error) {
	fields := sanitizeFields(req.Fields)
	var rows []struct {
		Code   string
		Name   string
		Type   string
		Status string
	}
	if err := s.deps.DB.WithContext(ctx).
		Table("product_spus").
		Select("code,name,type,status").
		Where("tenant_uuid = ?", ctxTenant(ctx)).
		Limit(500).
		Scan(&rows).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	if err := writer.Write(fields); err != nil {
		return "", err
	}
	for _, row := range rows {
		record := make([]string, len(fields))
		for i, field := range fields {
			switch field {
			case "code":
				record[i] = row.Code
			case "name":
				record[i] = row.Name
			case "type":
				record[i] = row.Type
			case "status":
				record[i] = row.Status
			default:
				record[i] = ""
			}
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	path, err := persistSnapshot(fmt.Sprintf("spu-export-%s.csv", taskID), buffer.Bytes())
	if err != nil {
		return "", err
	}
	return path, nil
}

func sanitizeFields(fields []string) []string {
	if len(fields) == 0 {
		return []string{"code", "name", "type", "status"}
	}
	var sanitized []string
	for _, field := range fields {
		key := strings.ToLower(strings.TrimSpace(field))
		if key == "" {
			continue
		}
		sanitized = append(sanitized, key)
	}
	if len(sanitized) == 0 {
		return []string{"code", "name", "type", "status"}
	}
	return sanitized
}

func parseImportRecords(payload []byte) ([]ImportRecord, error) {
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("解析表头失败: %w", err)
	}
	index := mapColumns(header)
	if _, ok := index["code"]; !ok {
		return nil, errors.New("模板缺少 code 列")
	}
	if _, ok := index["name"]; !ok {
		return nil, errors.New("模板缺少 name 列")
	}
	line := 1
	var records []ImportRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("解析第 %d 行失败: %w", line+1, err)
		}
		line++
		if isEmptyRow(row) {
			continue
		}
		skuIndex := -1
		if idx, ok := index["sku_codes"]; ok {
			skuIndex = idx
		}
		record := ImportRecord{
			Row:      line,
			Code:     pick(row, index["code"]),
			Name:     pick(row, index["name"]),
			SKUCodes: splitSKUCodes(pick(row, skuIndex)),
		}
		records = append(records, record)
	}
	return records, nil
}

func mapColumns(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, column := range header {
		key := strings.ToLower(strings.TrimSpace(column))
		if key != "" {
			index[key] = i
		}
	}
	return index
}

func pick(row []string, idx int) string {
	if idx < 0 {
		return ""
	}
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func splitSKUCodes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func persistFailureReport(taskID string, failures []ImportFailure) (string, error) {
	if len(failures) == 0 {
		return "", nil
	}
	body, err := json.MarshalIndent(failures, "", "  ")
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("spu-import-report-%s.json", taskID)
	return persistSnapshot(filename, body)
}

func persistSnapshot(filename string, content []byte) (string, error) {
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		return "", err
	}
	path := filepath.Join("tmp", filename)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *ImportService) updateExportTaskStatus(ctx context.Context, taskID, status, download string) error {
	updates := map[string]any{
		"status":       status,
		"download_url": download,
	}
	return s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPUExportTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func (s *ImportService) emitAudit(ctx context.Context, tenantID, taskID, event string, payload map[string]any) {
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return
	}
	body, _ := json.Marshal(payload)
	entry := productmodel.SPUAuditLog{
		ID:         uuid.NewString(),
		TenantUUID: tenantID,
		SPUID:      uuidString(),
		EventType:  event,
		Payload:    datatypes.JSON(body),
		Operator:   actorFromContext(ctx),
	}
	_ = s.deps.DB.WithContext(ctx).Create(&entry).Error
}

func (s *ImportService) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil && s.deps != nil {
		ctx = s.deps.Ctx
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return strings.TrimSpace(tid), nil
	}
	return "", ErrMissingTenant
}

func encodeMetadata(meta map[string]any) datatypes.JSON {
	if len(meta) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	body, err := json.Marshal(meta)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(body)
}

func ctxTenant(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok {
		return tid
	}
	return ""
}
