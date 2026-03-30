package customer

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

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	taskcenter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	taskbusx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/taskbus"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service 提供客户域的占位实现，依赖宿主 DB。
type Service struct {
	deps          *app.Deps
	repo          *customerrepo.Repository
	jobs          *taskcenter.Store
	taskSubmitter taskcenter.TaskSubmitter
	taskReporter  taskcenter.TaskReporter
}

// NewService 创建客户服务。
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("customer service requires initialized DB dependency")
	}
	frameworkTaskClient := taskcenter.NewFrameworkTaskClient("", "")
	return &Service{
		deps:          deps,
		repo:          customerrepo.NewRepository(deps.DB),
		jobs:          taskcenter.DefaultStore(),
		taskSubmitter: frameworkTaskClient,
		taskReporter:  frameworkTaskClient,
	}
}

// DB exposes the underlying DB handle for read-only feature extensions.
func (s *Service) DB() *gorm.DB {
	if s == nil || s.deps == nil {
		return nil
	}
	return s.deps.DB
}

// ListFilters 描述查询条件。
type ListFilters struct {
	Keyword   string
	Tier      string
	Source    string
	Region    string
	RiskLevel string
	Type      string
	Tags      []string
	Page      int
	PageSize  int
	Sort      string
}

// CustomerListResult 列表响应。
type CustomerListResult struct {
	Data []Customer       `json:"data"`
	Meta CustomerListMeta `json:"meta"`
}

// GetCustomer returns detail of a single customer by business ID.
func (s *Service) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	entity, err := s.repo.FindByCustomerID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrCustomerNotFound
	}
	result := convertEntityToCustomer(entity)
	if entity.LastOrderAt == nil {
		if lastAt, lastAmount, err := s.fetchLastOrderSnapshot(ctx, entity.CustomerID); err == nil && lastAt != nil {
			result.LastOrderAt = lastAt.UTC().Format(time.RFC3339)
			result.LastOrderAmount = lastAmount
		}
	}
	return &result, nil
}

// CustomerListMeta 附带分页信息。
type CustomerListMeta struct {
	Total       int     `json:"total"`
	SavedViewID *string `json:"savedViewId,omitempty"`
}

// Customer 兼容前端所需字段。
type Customer struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Type                string              `json:"type"`
	Email               string              `json:"email,omitempty"`
	Phone               string              `json:"phone,omitempty"`
	Country             string              `json:"country,omitempty"`
	Region              string              `json:"region,omitempty"`
	MembershipTier      string              `json:"membershipTier,omitempty"`
	MembershipTierLabel string              `json:"membershipTierLabel,omitempty"`
	GrowthValue         int                 `json:"growthValue,omitempty"`
	Points              int                 `json:"points,omitempty"`
	LastOrderAmount     float64             `json:"lastOrderAmount,omitempty"`
	LastOrderAt         string              `json:"lastOrderAt,omitempty"`
	Status              string              `json:"status,omitempty"`
	RiskLevel           string              `json:"riskLevel,omitempty"`
	Source              string              `json:"source,omitempty"`
	AccountManager      string              `json:"accountManager,omitempty"`
	Tags                []string            `json:"tags,omitempty"`
	CreatedAt           string              `json:"createdAt,omitempty"`
	UpdatedAt           string              `json:"updatedAt,omitempty"`
	Notes               string              `json:"notes,omitempty"`
	MaskedFields        []string            `json:"maskedFields,omitempty"`
	MembershipSnapshot  *MembershipSnapshot `json:"membershipSnapshot,omitempty"`
	Metadata            map[string]any      `json:"metadata,omitempty"`
}

// MembershipSnapshot 快照。
type MembershipSnapshot struct {
	CustomerID        string              `json:"customerId"`
	Tier              string              `json:"tier"`
	GrowthValue       int                 `json:"growthValue"`
	Points            int                 `json:"points"`
	RetentionStatus   string              `json:"retentionStatus"`
	Benefits          []MembershipBenefit `json:"benefits,omitempty"`
	LastBenefitUsedAt *string             `json:"lastBenefitUsedAt,omitempty"`
}

// MembershipBenefit 权益状态。
type MembershipBenefit struct {
	Name string `json:"name"`
	Used bool   `json:"used"`
}

// ListCustomers 返回过滤后的客户列表。
func (s *Service) ListCustomers(ctx context.Context, filters ListFilters) (*CustomerListResult, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	sanitized := filters
	sanitized.Tags = normalizeTags(filters.Tags)
	sanitized.Page = page
	sanitized.PageSize = pageSize
	if strings.TrimSpace(sanitized.Sort) == "" && strings.TrimSpace(sanitized.Keyword) == "" {
		sanitized.Sort = "-lastOrderAt"
	}

	repoFilters := customerrepo.ListQueryOptions{
		Keyword:   sanitized.Keyword,
		Tier:      sanitized.Tier,
		Source:    sanitized.Source,
		Region:    sanitized.Region,
		RiskLevel: sanitized.RiskLevel,
		Type:      sanitized.Type,
		Tags:      sanitized.Tags,
		Sort:      sanitized.Sort,
		Page:      sanitized.Page,
		PageSize:  sanitized.PageSize,
	}

	pageResult, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}
	customers := make([]Customer, 0, len(pageResult.List))
	for _, entity := range pageResult.List {
		if entity == nil {
			continue
		}
		customers = append(customers, convertEntityToCustomer(entity))
	}
	return &CustomerListResult{
		Data: customers,
		Meta: CustomerListMeta{Total: int(pageResult.Total)},
	}, nil
}

func convertEntityToCustomer(entity *customermodel.Customer) Customer {
	if entity == nil {
		return Customer{}
	}
	result := Customer{
		ID:                  entity.CustomerID,
		Name:                entity.Name,
		Type:                entity.Type,
		Email:               entity.Email,
		Phone:               entity.Phone,
		Country:             entity.Country,
		Region:              entity.Region,
		MembershipTier:      entity.MembershipTier,
		MembershipTierLabel: entity.MembershipTierLabel,
		GrowthValue:         entity.GrowthValue,
		Points:              entity.Points,
		LastOrderAmount:     entity.LastOrderAmount,
		LastOrderAt:         formatTimePtr(entity.LastOrderAt),
		Status:              entity.Status,
		RiskLevel:           entity.RiskLevel,
		Source:              entity.Source,
		AccountManager:      entity.AccountManager,
		Tags:                decodeStringSlice(entity.Tags),
		MaskedFields:        decodeStringSlice(entity.MaskedFields),
		MembershipSnapshot:  decodeMembershipSnapshot(entity.MembershipSnapshot),
		Metadata:            decodeMetadata(entity.Metadata),
		Notes:               entity.Notes,
		CreatedAt:           entity.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           entity.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return result
}

func copyCustomerToEntity(entity *customermodel.Customer, c Customer) {
	entity.CustomerID = c.ID
	if entity.CustomerID == "" {
		entity.CustomerID = utils.NewUUID()
	}
	entity.Name = c.Name
	entity.Type = c.Type
	entity.Email = c.Email
	entity.Phone = c.Phone
	entity.Country = c.Country
	entity.Region = c.Region
	entity.MembershipTier = c.MembershipTier
	entity.MembershipTierLabel = c.MembershipTierLabel
	entity.GrowthValue = c.GrowthValue
	entity.Points = c.Points
	entity.LastOrderAmount = c.LastOrderAmount
	entity.LastOrderAt = parseTimePtr(c.LastOrderAt)
	entity.Status = c.Status
	entity.RiskLevel = c.RiskLevel
	entity.Source = c.Source
	entity.AccountManager = c.AccountManager
	entity.Tags = encodeStringSlice(c.Tags)
	entity.MaskedFields = encodeStringSlice(c.MaskedFields)
	entity.MembershipSnapshot = encodeMembershipSnapshot(c.MembershipSnapshot)
	entity.Metadata = encodeMetadata(c.Metadata)
	entity.Notes = c.Notes
	if created := parseTime(c.CreatedAt); !created.IsZero() {
		entity.CreatedAt = created
	}
	if updated := parseTime(c.UpdatedAt); !updated.IsZero() {
		entity.UpdatedAt = updated
	}
}

func (s *Service) fetchLastOrderSnapshot(ctx context.Context, customerID string) (*time.Time, float64, error) {
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return nil, 0, nil
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, 0, err
	}
	type row struct {
		CreatedAt   time.Time `gorm:"column:created_at"`
		TotalAmount int64     `gorm:"column:total_amount"`
	}
	var last row
	if err := s.deps.DB.WithContext(ctx).
		Model(&ordermodel.Order{}).
		Select("created_at, total_amount").
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, strings.TrimSpace(customerID)).
		Order("created_at DESC").
		Limit(1).
		Scan(&last).Error; err != nil {
		return nil, 0, err
	}
	if last.CreatedAt.IsZero() {
		return nil, 0, nil
	}
	return &last.CreatedAt, float64(last.TotalAmount) / 100.0, nil
}

const importJobType = "customer_import"
const exportJobType = "customer_export"
const importConflictReportURLPrefix = "/api/v1/admin/customers/import/conflicts/"

type ImportConflictStrategy string

const (
	ImportConflictStrategyFail ImportConflictStrategy = "fail"
	ImportConflictStrategySkip ImportConflictStrategy = "skip"
)

type ImportConflictRow struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
}

type ImportJobResult struct {
	Processed int                `json:"processed"`
	Imported  int                `json:"imported"`
	Conflicts []ImportConflictRow `json:"conflicts,omitempty"`
}

type ExportJobRequest struct {
	Filters map[string]any
	Fields  []string
}

// StartImportJob registers an async import task and returns task ID for polling.
func (s *Service) StartImportJob(ctx context.Context, filename string, payload []byte, strategy ImportConflictStrategy) (string, error) {
	if len(payload) == 0 {
		return "", fmt.Errorf("导入文件内容为空")
	}
	tenant := tenantFromContext(ctx)
	if tenant == "" {
		return "", authx.ErrTenantMissing
	}
	if strategy != ImportConflictStrategySkip {
		strategy = ImportConflictStrategyFail
	}
	meta := map[string]any{
		"tenantUuid":        tenant,
		"filename":          filename,
		"conflictStrategy":  string(strategy),
	}
	taskID, err := s.submitFrameworkTask(ctx, importJobType, tenant, meta)
	if err != nil {
		return "", err
	}
	go s.executeImportJob(authx.ContextWithTenantUUID(context.Background(), tenant), taskID, filename, payload, strategy)
	return taskID, nil
}

// StartExportJob registers an async export task and returns task ID for tracking.
func (s *Service) StartExportJob(ctx context.Context, req ExportJobRequest) (string, error) {
	tenant := tenantFromContext(ctx)
	if tenant == "" {
		return "", authx.ErrTenantMissing
	}
	fields := sanitizeExportFields(req.Fields)
	meta := map[string]any{
		"tenantUuid": tenant,
		"fields":     fields,
	}
	taskID, err := s.submitFrameworkTask(ctx, exportJobType, tenant, meta)
	if err != nil {
		return "", err
	}
	go s.executeExportJob(authx.ContextWithTenantUUID(context.Background(), tenant), taskID, ExportJobRequest{Filters: req.Filters, Fields: fields})
	return taskID, nil
}

func (s *Service) executeExportJob(ctx context.Context, taskID string, req ExportJobRequest) {
	runningMessage := "正在导出客户"
	_, _ = s.jobs.SetStatus(taskID, "running", runningMessage)
	s.reportFrameworkTask(ctx, taskID, "running", runningMessage, "", nil)
	s.publishTaskEvent(ctx, taskID, "running", runningMessage, "", 0)

	rows, err := s.loadExportRows(ctx, req)
	if err != nil {
		_, _ = s.jobs.Fail(taskID, err)
		completedAt := time.Now().UTC()
		s.reportFrameworkTask(ctx, taskID, "failed", err.Error(), "", &completedAt)
		s.publishTaskEvent(ctx, taskID, "failed", err.Error(), "", 100)
		return
	}

	path, err := s.persistExportCSV(taskID, req.Fields, rows, func(processed, total int) {
		if total <= 0 {
			return
		}
		progress := int(float64(processed) * 100 / float64(total))
		s.publishTaskEvent(ctx, taskID, "running", runningMessage, "", clampProgress(progress))
	})
	if err != nil {
		_, _ = s.jobs.Fail(taskID, err)
		completedAt := time.Now().UTC()
		s.reportFrameworkTask(ctx, taskID, "failed", err.Error(), "", &completedAt)
		s.publishTaskEvent(ctx, taskID, "failed", err.Error(), "", 100)
		return
	}

	message := "导出完成，可下载文件"
	_, _ = s.jobs.Success(taskID, message, path)
	completedAt := time.Now().UTC()
	s.reportFrameworkTask(ctx, taskID, "success", message, path, &completedAt)
	s.publishTaskEvent(ctx, taskID, "success", message, path, 100)
}
func (s *Service) executeImportJob(ctx context.Context, taskID, filename string, payload []byte, strategy ImportConflictStrategy) {
	runningMessage := fmt.Sprintf("正在导入 %s", filename)
	_, _ = s.jobs.SetStatus(taskID, "running", runningMessage)
	s.reportFrameworkTask(ctx, taskID, "running", runningMessage, "", nil)
	totalRows := countImportDataRows(payload)
	s.publishTaskEvent(ctx, taskID, "running", runningMessage, "", 0)
	result, err := s.importCustomersFromCSV(ctx, bytes.NewReader(payload), strategy, func(processed int) {
		if totalRows <= 0 {
			return
		}
		progress := int(float64(processed) * 100 / float64(totalRows))
		s.publishTaskEvent(ctx, taskID, "running", runningMessage, "", clampProgress(progress))
	})
	if err != nil {
		reportURL := ""
		if len(result.Conflicts) > 0 {
			if path, reportErr := s.persistImportConflictCSV(taskID, result.Conflicts); reportErr == nil {
				reportURL = path
			}
		}
		_, _ = s.jobs.Fail(taskID, err)
		completedAt := time.Now().UTC()
		s.reportFrameworkTask(ctx, taskID, "failed", err.Error(), reportURL, &completedAt)
		s.publishTaskEvent(ctx, taskID, "failed", err.Error(), reportURL, 100)
		return
	}

	reportURL := ""
	if len(result.Conflicts) > 0 {
		if path, reportErr := s.persistImportConflictCSV(taskID, result.Conflicts); reportErr == nil {
			reportURL = path
		}
	}

	message := fmt.Sprintf("成功导入 %d 条客户", result.Imported)
	if len(result.Conflicts) > 0 {
		message = fmt.Sprintf("成功导入 %d 条客户，冲突 %d 条", result.Imported, len(result.Conflicts))
	}
	_, _ = s.jobs.Success(taskID, message, reportURL)
	completedAt := time.Now().UTC()
	s.reportFrameworkTask(ctx, taskID, "success", message, reportURL, &completedAt)
	s.publishTaskEvent(ctx, taskID, "success", message, reportURL, 100)
}

func ImportConflictReportPath(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ""
	}
	return filepath.Join("tmp", fmt.Sprintf("customer-import-conflicts-%s.csv", taskID))
}

func ImportConflictReportURL(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ""
	}
	return importConflictReportURLPrefix + taskID
}

func (s *Service) submitFrameworkTask(ctx context.Context, taskType, tenantUUID string, metadata map[string]any) (string, error) {
	if s == nil {
		return "", fmt.Errorf("customer service unavailable")
	}
	if s.taskSubmitter == nil {
		return s.createLocalTask(taskType, tenantUUID, metadata)
	}
	result, err := s.taskSubmitter.Submit(ctx, taskcenter.TaskSubmitRequest{
		Type:       taskType,
		TenantUUID: tenantUUID,
		Metadata:   metadata,
	})
	if err != nil {
		if errors.Is(err, taskcenter.ErrTaskSubmitNotAvailable) {
			return s.createLocalTask(taskType, tenantUUID, metadata)
		}
		return "", fmt.Errorf("提交 framework 任务失败: %w", err)
	}
	if result == nil || strings.TrimSpace(result.TaskID) == "" {
		return "", fmt.Errorf("framework task id 为空")
	}
	return strings.TrimSpace(result.TaskID), nil
}

func (s *Service) createLocalTask(taskType, tenantUUID string, metadata map[string]any) (string, error) {
	if s == nil || s.jobs == nil {
		return "", fmt.Errorf("无法创建本地任务")
	}
	localMeta := map[string]any{"tenantUuid": strings.TrimSpace(tenantUUID)}
	for key, value := range metadata {
		localMeta[key] = value
	}
	job := s.jobs.Create(taskType, localMeta)
	if job == nil || strings.TrimSpace(job.TaskID) == "" {
		return "", fmt.Errorf("无法创建本地任务")
	}
	return strings.TrimSpace(job.TaskID), nil
}

func (s *Service) reportFrameworkTask(ctx context.Context, taskID, status, message, downloadURL string, completedAt *time.Time) {
	if s == nil || s.taskReporter == nil {
		return
	}
	tenantUUID := tenantFromContext(ctx)
	_ = s.taskReporter.Update(ctx, taskcenter.TaskUpdateRequest{
		TaskID:      taskID,
		TenantUUID:  tenantUUID,
		Status:      status,
		Message:     message,
		DownloadURL: downloadURL,
		CompletedAt: completedAt,
	})
}

func (s *Service) publishTaskEvent(ctx context.Context, taskID, status, message, downloadURL string, progress int) {
	if s == nil || s.deps == nil || s.deps.TaskBus == nil {
		return
	}
	tenantUUID := tenantFromContext(ctx)
	payload := map[string]any{
		"taskId":      taskID,
		"task_id":     taskID,
		"id":          taskID,
		"status":      status,
		"message":     message,
		"tenant_uuid": tenantUUID,
		"progress":    clampProgress(progress),
	}
	if strings.TrimSpace(downloadURL) != "" {
		payload["downloadUrl"] = downloadURL
		payload["download_url"] = downloadURL
	}
	metadata := map[string]string{"tenant_uuid": tenantUUID}
	topics := []string{"task.progress", "powerx.task.progress.v1", "worker.task.updated"}
	for _, topic := range topics {
		_ = s.deps.TaskBus.Publish(ctx, taskbusx.Event{Topic: topic, Payload: payload, Metadata: metadata})
	}
}

func (s *Service) importCustomersFromCSV(ctx context.Context, reader io.Reader, strategy ImportConflictStrategy, onProgress func(processed int)) (ImportJobResult, error) {
	result := ImportJobResult{}
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	header, err := csvReader.Read()
	if err != nil {
		return result, fmt.Errorf("解析表头失败: %w", err)
	}
	index := map[string]int{}
	for i, column := range header {
		key := strings.ToLower(strings.TrimSpace(column))
		index[key] = i
	}
	required := []string{"name", "type", "phone", "email", "source", "membershiptier"}
	for _, key := range required {
		if _, ok := index[key]; !ok {
			return result, fmt.Errorf("模板缺少必填列: %s", key)
		}
	}
	line := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, fmt.Errorf("解析第 %d 行失败: %w", line+1, err)
		}
		line++
		if isEmptyRecord(record) {
			continue
		}
		result.Processed++
		input := s.recordToInput(record, index)
		if _, err := s.CreateCustomer(ctx, input); err != nil {
			conflict := ImportConflictRow{
				Row:    line,
				Reason: err.Error(),
				Name:   input.Name,
				Email:  input.Email,
				Phone:  input.Phone,
			}
			if errors.Is(err, ErrCustomerConflict) {
				result.Conflicts = append(result.Conflicts, conflict)
				if onProgress != nil {
					onProgress(result.Processed)
				}
				if strategy == ImportConflictStrategySkip {
					continue
				}
				return result, fmt.Errorf("第 %d 行导入失败: %v", line, err)
			}
			if onProgress != nil {
				onProgress(result.Processed)
			}
			return result, fmt.Errorf("第 %d 行导入失败: %v", line, err)
		}
		result.Imported++
		if onProgress != nil {
			onProgress(result.Processed)
		}
	}
	return result, nil
}

func (s *Service) persistImportConflictCSV(taskID string, conflicts []ImportConflictRow) (string, error) {
	if len(conflicts) == 0 {
		return "", nil
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return "", fmt.Errorf("task id is required")
	}
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		return "", err
	}
	path := ImportConflictReportPath(taskID)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"row", "reason", "name", "email", "phone"}); err != nil {
		return "", err
	}
	for _, item := range conflicts {
		record := []string{
			fmt.Sprintf("%d", item.Row),
			item.Reason,
			item.Name,
			item.Email,
			item.Phone,
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return ImportConflictReportURL(taskID), nil
}

func toString(v any) string {
	if value, ok := v.(string); ok {
		return value
	}
	return ""
}

func sanitizeExportFields(fields []string) []string {
	if len(fields) == 0 {
		return []string{"id", "name", "type", "phone", "email", "source", "membershipTier", "status", "createdAt"}
	}
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		key := strings.TrimSpace(field)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	if len(out) == 0 {
		return []string{"id", "name", "type", "phone", "email", "source", "membershipTier", "status", "createdAt"}
	}
	return out
}

func (s *Service) loadExportRows(ctx context.Context, req ExportJobRequest) ([]Customer, error) {
	tenant := tenantFromContext(ctx)
	if tenant == "" {
		return nil, authx.ErrTenantMissing
	}
	query := s.deps.DB.WithContext(ctx).
		Model(&customermodel.Customer{}).
		Where("tenant_uuid = ?", tenant).
		Order("created_at DESC").
		Limit(5000)

	if keyword := strings.TrimSpace(toString(req.Filters["keyword"])); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ?", like, like, like)
	}

	var entities []customermodel.Customer
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	rows := make([]Customer, 0, len(entities))
	for i := range entities {
		rows = append(rows, convertEntityToCustomer(&entities[i]))
	}
	return rows, nil
}

func (s *Service) persistExportCSV(taskID string, fields []string, rows []Customer, onProgress func(processed, total int)) (string, error) {
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		return "", err
	}
	path := filepath.Join("tmp", fmt.Sprintf("customer-export-%s.csv", taskID))
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(fields); err != nil {
		return "", err
	}
	total := len(rows)
	for i, row := range rows {
		record := make([]string, len(fields))
		for idx, field := range fields {
			record[idx] = exportFieldValue(row, field)
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
		if onProgress != nil {
			onProgress(i+1, total)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return path, nil
}

func exportFieldValue(row Customer, field string) string {
	switch strings.TrimSpace(field) {
	case "id":
		return row.ID
	case "name":
		return row.Name
	case "type":
		return row.Type
	case "phone":
		return row.Phone
	case "email":
		return row.Email
	case "country":
		return row.Country
	case "region":
		return row.Region
	case "source":
		return row.Source
	case "membershipTier":
		return row.MembershipTier
	case "membershipTierLabel":
		return row.MembershipTierLabel
	case "status":
		return row.Status
	case "riskLevel":
		return row.RiskLevel
	case "accountManager":
		return row.AccountManager
	case "createdAt":
		return row.CreatedAt
	case "updatedAt":
		return row.UpdatedAt
	default:
		return ""
	}
}

func countImportDataRows(payload []byte) int {
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.TrimLeadingSpace = true
	if _, err := reader.Read(); err != nil {
		return 0
	}
	total := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return total
		}
		if isEmptyRecord(record) {
			continue
		}
		total++
	}
	return total
}

func clampProgress(progress int) int {
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}
	return progress
}

func (s *Service) recordToInput(record []string, index map[string]int) CreateCustomerInput {
	read := func(key string) string {
		i, ok := index[key]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}
	tags := normalizeTags(splitCSVTags(read("tags")))
	return CreateCustomerInput{
		Name:           read("name"),
		Type:           read("type"),
		Email:          read("email"),
		Phone:          read("phone"),
		Source:         fallbackString(read("source"), "import"),
		Country:        read("country"),
		Region:         read("region"),
		MembershipTier: fallbackString(read("membershiptier"), "bronze"),
		AccountManager: read("accountmanager"),
		Tags:           tags,
		Notes:          read("notes"),
	}
}

func splitCSVTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var tags []string
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func encodeStringSlice(values []string) datatypes.JSON {
	if len(values) == 0 {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeStringSlice(data datatypes.JSON) []string {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func encodeMembershipSnapshot(snapshot *MembershipSnapshot) datatypes.JSON {
	if snapshot == nil {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeMembershipSnapshot(data datatypes.JSON) *MembershipSnapshot {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var snapshot MembershipSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil
	}
	return &snapshot
}

func encodeMetadata(meta map[string]any) datatypes.JSON {
	if len(meta) == 0 {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeMetadata(data datatypes.JSON) map[string]any {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil
	}
	return meta
}

func parseTimePtr(value string) *time.Time {
	t := parseTime(value)
	if t.IsZero() {
		return nil
	}
	return &t
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

func normalizeTags(tags []string) []string {
	var out []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if strings.Contains(tag, ",") {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				p := strings.TrimSpace(part)
				if p != "" {
					out = append(out, p)
				}
			}
			continue
		}
		out = append(out, tag)
	}
	return out
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
