package spu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/domain/repository/product"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/jobs/channels"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ErrMissingTenant indicates the request context did not carry tenant metadata.
var ErrMissingTenant = errors.New("tenant context missing")

// ErrInvalidStatus indicates an unsupported state transition.
var ErrInvalidStatus = errors.New("invalid status transition")

// Service orchestrates tenant-scoped SPU lifecycle operations.
type Service struct {
	deps         *app.Deps
	spuRepo      *productrepo.SPURepository
	versionRepo  *productrepo.VersionRepository
	channelRepo  *productrepo.ChannelRepository
	planRepo     *productrepo.SubscriptionPlanRepository
	approvalRepo *productrepo.ApprovalRepository
	locales      *LocaleService
	publisher    channels.Publisher
	metrics      *productmetrics.SPUMetrics
}

// NewService wires repositories and helpers using the shared dependency bundle.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("spu service requires initialized DB dependency")
	}
	logger := deps.RuntimeLogger(nil, "product-spu-publisher", nil)
	return &Service{
		deps:         deps,
		spuRepo:      productrepo.NewSPURepository(deps.DB),
		versionRepo:  productrepo.NewVersionRepository(deps.DB),
		channelRepo:  productrepo.NewChannelRepository(deps.DB),
		planRepo:     productrepo.NewSubscriptionPlanRepository(deps.DB),
		approvalRepo: productrepo.NewApprovalRepository(deps.DB),
		locales:      NewLocaleService([]string{"zh-CN"}),
		publisher:    channels.NewAsyncPublisher(logger),
		metrics:      resolveSPUMetrics(deps, "product-spu-service"),
	}
}

// UpsertSPURequest captures fields required to create or update an SPU draft.
type UpsertSPURequest struct {
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	CategoryID      string          `json:"categoryId"`
	CategoryPath    string          `json:"categoryPath"`
	BrandID         string          `json:"brandId"`
	DefaultLocale   string          `json:"defaultLocale"`
	Tags            []string        `json:"tags"`
	ResponsibleUser string          `json:"responsibleUser"`
	Locales         []LocaleContent `json:"locales"`
}

// Normalize trims whitespace and deduplicates tags for consistent processing.
func (r *UpsertSPURequest) Normalize() {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	r.Type = strings.TrimSpace(strings.ToLower(r.Type))
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.CategoryPath = strings.TrimSpace(r.CategoryPath)
	r.BrandID = strings.TrimSpace(r.BrandID)
	r.DefaultLocale = strings.TrimSpace(r.DefaultLocale)
	r.ResponsibleUser = strings.TrimSpace(r.ResponsibleUser)
	tagSet := map[string]struct{}{}
	for _, tag := range r.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			tagSet[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	r.Tags = make([]string, 0, len(tagSet))
	for tag := range tagSet {
		r.Tags = append(r.Tags, tag)
	}
}

// ValidateUpsertRequest enforces field level constraints before persistence.
func (s *Service) ValidateUpsertRequest(ctx context.Context, req UpsertSPURequest) error {
	if s == nil {
		return errors.New("spu service not initialized")
	}
	req.Normalize()
	req.Locales = s.locales.Normalize(req.Locales)
	var errs ValidationErrors
	if req.Code == "" {
		errs = errs.add("code", "code is required")
	} else if len(req.Code) > 120 {
		errs = errs.add("code", "code cannot exceed 120 characters")
	}
	if req.Name == "" {
		errs = errs.add("name", "name is required")
	}
	supportedTypes := map[string]struct{}{"one_time": {}, "subscription": {}, "bundle": {}}
	if req.Type == "" {
		errs = errs.add("type", "type is required")
	} else if _, ok := supportedTypes[req.Type]; !ok {
		errs = errs.add("type", "unsupported type")
	}
	if req.CategoryID == "" {
		errs = errs.add("categoryId", "category id is required")
	}
	if req.CategoryPath == "" {
		errs = errs.add("categoryPath", "category path is required")
	}
	if req.DefaultLocale == "" {
		errs = errs.add("defaultLocale", "default locale is required")
	}
	localeErrs := s.locales.Validate(req.DefaultLocale, req.Locales)
	if !localeErrs.empty() {
		errs = append(errs, localeErrs...)
	}
	if errs.empty() {
		if _, err := s.tenantFromContext(ctx); err != nil {
			return err
		}
		return nil
	}
	return errs
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = s.deps.Ctx
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", ErrMissingTenant
}

// ListFilters describes query params for listing SPUs.
type ListFilters struct {
	Keyword  string
	Status   string
	Type     string
	Page     int
	PageSize int
}

// ListResult wraps paginated SPU summaries.
type ListResult struct {
	Items    []SPUSummary `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// SPUSummary contains lightweight information for table rendering.
type SPUSummary struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SPUDetail exposes fields consumed by the edit/detail screens.
type SPUDetail struct {
	ID             string          `json:"id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	CategoryID     string          `json:"categoryId"`
	CategoryPath   string          `json:"categoryPath"`
	BrandID        string          `json:"brandId,omitempty"`
	DefaultLocale  string          `json:"defaultLocale"`
	Status         string          `json:"status"`
	Tags           []string        `json:"tags"`
	Responsible    string          `json:"responsibleUser,omitempty"`
	Locales        []LocaleContent `json:"locales,omitempty"`
	CurrentVersion string          `json:"currentVersionId,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// List returns paginated SPU summaries filtered by keyword/status.
func (s *Service) List(ctx context.Context, filters ListFilters) (*ListResult, error) {
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	query := s.deps.DB.WithContext(ctx).Model(&productmodel.SPU{}).
		Where("tenant_uuid = ?", tenantID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Keyword != "" {
		like := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var records []productmodel.SPU
	if err := query.Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]SPUSummary, 0, len(records))
	for _, rec := range records {
		items = append(items, SPUSummary{
			ID:        rec.ID,
			Code:      rec.Code,
			Name:      rec.Name,
			Type:      rec.Type,
			Status:    rec.Status,
			UpdatedAt: rec.UpdatedAt,
		})
	}
	return &ListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// CreateDraft persists a new SPU draft and associated version snapshot.
func (s *Service) CreateDraft(ctx context.Context, req UpsertSPURequest) (*SPUDetail, error) {
	if err := s.ValidateUpsertRequest(ctx, req); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Normalize()
	req.Locales = s.locales.Normalize(req.Locales)
	payload := map[string]any{
		"input":   req,
		"version": "draft",
		"skus":    []any{},
	}
	body, _ := json.Marshal(payload)

	result := &SPUDetail{}
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		now := time.Now().UTC()
		spu := &productmodel.SPU{
			ID:              uuidString(),
			TenantUUID:      tenantID,
			Code:            req.Code,
			Name:            req.Name,
			Type:            req.Type,
			CategoryID:      req.CategoryID,
			CategoryPath:    req.CategoryPath,
			BrandID:         req.BrandID,
			DefaultLocale:   req.DefaultLocale,
			Status:          "draft",
			ResponsibleUser: req.ResponsibleUser,
			Tags:            pqStringArray(req.Tags),
			ChannelsSummary: datatypes.JSON([]byte(`{}`)),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := tx.Create(spu).Error; err != nil {
			return err
		}
		version := &productmodel.SPUVersion{
			ID:            uuidString(),
			TenantUUID:    tenantID,
			SPUID:         spu.ID,
			VersionNumber: 1,
			Status:        "draft",
			Payload:       datatypes.JSON(body),
			SubmittedBy:   req.ResponsibleUser,
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		if err := tx.Model(spu).Update("current_version_id", version.ID).Error; err != nil {
			return err
		}
		result = convertToDetail(spu, req.Locales, version.ID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateDraft persists changes to an existing draft SPU.
func (s *Service) UpdateDraft(ctx context.Context, id string, req UpsertSPURequest) (*SPUDetail, error) {
	if err := s.ValidateUpsertRequest(ctx, req); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Normalize()
	req.Locales = s.locales.Normalize(req.Locales)
	var result *SPUDetail
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		if strings.ToLower(spu.Status) != "draft" {
			return fmt.Errorf("only draft SPU can be updated (current: %s)", spu.Status)
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"name":             req.Name,
			"type":             req.Type,
			"category_id":      req.CategoryID,
			"category_path":    req.CategoryPath,
			"brand_id":         req.BrandID,
			"default_locale":   req.DefaultLocale,
			"responsible_user": req.ResponsibleUser,
			"tags":             pqStringArray(req.Tags),
			"updated_at":       now,
		}
		if err := tx.Model(&spu).Updates(updates).Error; err != nil {
			return err
		}
		spu.Name = req.Name
		spu.Type = req.Type
		spu.CategoryID = req.CategoryID
		spu.CategoryPath = req.CategoryPath
		spu.BrandID = req.BrandID
		spu.DefaultLocale = req.DefaultLocale
		spu.ResponsibleUser = req.ResponsibleUser
		spu.Tags = pqStringArray(req.Tags)
		spu.UpdatedAt = now
		versionID := derefString(spu.CurrentVersionID)
		payload := map[string]any{}
		if versionID != "" {
			var existingVersion productmodel.SPUVersion
			if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, versionID).First(&existingVersion).Error; err != nil {
				return err
			}
			payload = decodeVersionPayload(existingVersion.Payload)
		}
		payload["input"] = req
		if _, ok := payload["skus"]; !ok {
			payload["skus"] = []any{}
		}
		payload["version"] = "draft"
		body := encodeVersionPayload(payload)
		if versionID == "" {
			versionID = uuidString()
			version := &productmodel.SPUVersion{
				ID:            versionID,
				TenantUUID:    tenantID,
				SPUID:         spu.ID,
				VersionNumber: 1,
				Status:        "draft",
				Payload:       body,
			}
			if err := tx.Create(version).Error; err != nil {
				return err
			}
			if err := tx.Model(&spu).Update("current_version_id", versionID).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&productmodel.SPUVersion{}).
				Where("tenant_uuid = ? AND id = ?", tenantID, versionID).
				Updates(map[string]any{"payload": body, "status": "draft", "diff_summary": datatypes.JSON([]byte("{}"))}).Error; err != nil {
				return err
			}
		}
		result = convertToDetail(&spu, req.Locales, versionID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SubmitRequest describes payload for moving a draft into review.
type SubmitRequest struct {
	VersionID string `json:"versionId"`
	Comment   string `json:"comment"`
}

// PublishRequest describes payload for publishing a reviewed SPU.
type PublishRequest struct {
	VersionID   string   `json:"versionId"`
	Channels    []string `json:"channels"`
	PublishMode string   `json:"publishMode"`
}

// WithdrawRequest describes payload for deactivating channels.
type WithdrawRequest struct {
	Channels   []string `json:"channels"`
	WithdrawAt string   `json:"withdrawAt"`
	Reason     string   `json:"reason"`
}

// DeleteRequest captures audit info when removing an SPU.
type DeleteRequest struct {
	Reason string `json:"reason"`
}

// Submit transitions a draft SPU into reviewing state.
func (s *Service) Submit(ctx context.Context, id string, req SubmitRequest) (*SPUDetail, error) {
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	submitter := actorFromContext(ctx)
	var result *SPUDetail
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		if spu.Status != "draft" {
			return fmt.Errorf("spu status %s cannot submit", spu.Status)
		}
		versionID := req.VersionID
		if versionID == "" {
			versionID = derefString(spu.CurrentVersionID)
		}
		if versionID == "" {
			return errors.New("missing draft version")
		}
		var version productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, versionID).First(&version).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":       "reviewing",
			"submitted_at": now,
		}
		if submitter != "" && submitter != "system" {
			updates["submitted_by"] = submitter
		}
		if err := tx.Model(&version).Updates(updates).Error; err != nil {
			return err
		}
		version.Status = "reviewing"
		version.SubmittedAt = &now
		if submitter != "" && submitter != "system" {
			version.SubmittedBy = submitter
		}
		spu.Status = "reviewing"
		spu.UpdatedAt = now
		if err := tx.Model(&spu).Updates(map[string]any{"status": "reviewing", "updated_at": now}).Error; err != nil {
			return err
		}
		if err := s.ensureApprovalChain(tx, tenantID, &version, submitter); err != nil {
			return err
		}
		result = convertToDetail(&spu, nil, versionID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Publish marks a reviewed SPU as published and triggers channel tasks.
func (s *Service) Publish(ctx context.Context, id string, req PublishRequest) (*SPUDetail, error) {
	var vErrs ValidationErrors
	if strings.TrimSpace(req.VersionID) == "" {
		vErrs = vErrs.add("versionId", "versionId is required")
	}
	if len(req.Channels) == 0 {
		vErrs = vErrs.add("channels", "at least one channel is required")
	}
	if !vErrs.empty() {
		return nil, vErrs
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Channels = dedupeStrings(req.Channels)
	var (
		detail   *SPUDetail
		leadTime time.Duration
	)
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		if spu.Status != "reviewing" && spu.Status != "draft" {
			return fmt.Errorf("spu status %s cannot publish", spu.Status)
		}
		var version productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, req.VersionID, spu.ID).First(&version).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		taskID := ""
		if s.publisher != nil {
			task, err := s.publisher.EnqueuePublish(ctx, tenantID, spu.ID, version.ID, req.Channels)
			if err != nil {
				return err
			}
			taskID = task.TaskID
		}
		if err := tx.Model(&version).Updates(map[string]any{"status": "published", "approved_at": now}).Error; err != nil {
			return err
		}
		summary := buildChannelsSummary(req.Channels, "publishing", taskID)
		updates := map[string]any{
			"status":             "published",
			"current_version_id": version.ID,
			"channels_summary":   summary,
			"updated_at":         now,
		}
		if err := tx.Model(&spu).Updates(updates).Error; err != nil {
			return err
		}
		spu.Status = "published"
		spu.UpdatedAt = now
		spu.ChannelsSummary = summary
		spu.CurrentVersionID = &version.ID
		detail = convertToDetail(&spu, nil, version.ID)
		if !spu.CreatedAt.IsZero() {
			leadTime = now.Sub(spu.CreatedAt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.metrics != nil && leadTime > 0 {
		s.metrics.ObserveLeadTime(leadTime)
	}
	return detail, nil
}

// Withdraw marks specified channels as offboarded and optionally schedules the operation.
func (s *Service) Withdraw(ctx context.Context, id string, req WithdrawRequest) (*SPUDetail, error) {
	if s == nil {
		return nil, errors.New("spu service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	req.WithdrawAt = strings.TrimSpace(req.WithdrawAt)
	req.Channels = dedupeStrings(req.Channels)
	var vErrs ValidationErrors
	if req.Reason == "" {
		vErrs = vErrs.add("reason", "reason is required")
	}
	var scheduledAt *time.Time
	if req.WithdrawAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.WithdrawAt); err != nil {
			vErrs = vErrs.add("withdrawAt", "withdrawAt must be RFC3339 timestamp")
		} else {
			utc := parsed.UTC()
			scheduledAt = &utc
		}
	}
	if !vErrs.empty() {
		return nil, vErrs
	}
	var detail *SPUDetail
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		if spu.Status != "published" {
			return fmt.Errorf("spu status %s cannot withdraw", spu.Status)
		}
		var records []productmodel.ChannelVisibility
		if err := tx.Where("tenant_uuid = ? AND spu_id = ?", tenantID, spu.ID).Find(&records).Error; err != nil {
			return err
		}
		if len(records) == 0 {
			return errors.New("spu has no configured channels to withdraw")
		}
		recordMap := make(map[string]productmodel.ChannelVisibility, len(records))
		for _, record := range records {
			recordMap[strings.ToLower(record.Channel)] = record
		}
		targetChannels := make([]productmodel.ChannelVisibility, 0, len(records))
		if len(req.Channels) == 0 {
			targetChannels = records
		} else {
			for _, ch := range req.Channels {
				record, ok := recordMap[strings.ToLower(ch)]
				if !ok {
					return fmt.Errorf("channel %s not found on spu", ch)
				}
				targetChannels = append(targetChannels, record)
			}
		}
		if len(targetChannels) == 0 {
			return errors.New("no eligible channels selected for withdraw")
		}
		now := time.Now().UTC()
		effectiveAt := now
		if scheduledAt != nil {
			effectiveAt = *scheduledAt
		}
		feedback := map[string]any{
			"reason":    req.Reason,
			"operator":  actorFromContext(ctx),
			"action":    "withdraw",
			"timestamp": effectiveAt,
		}
		for _, record := range targetChannels {
			updates := map[string]any{
				"availability":  "offboarded",
				"withdraw_at":   effectiveAt,
				"updated_at":    now,
				"last_feedback": encodeGenericJSON(feedback),
			}
			if err := tx.Model(&productmodel.ChannelVisibility{}).
				Where("tenant_uuid = ? AND id = ?", tenantID, record.ID).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		var updated []productmodel.ChannelVisibility
		if err := tx.Where("tenant_uuid = ? AND spu_id = ?", tenantID, spu.ID).Find(&updated).Error; err != nil {
			return err
		}
		summary := summarizeChannelRecords(updated)
		if err := tx.Model(&productmodel.SPU{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, spu.ID).
			Update("channels_summary", summary).Error; err != nil {
			return err
		}
		var activeCount int64
		if err := tx.Model(&productmodel.ChannelVisibility{}).
			Where("tenant_uuid = ? AND spu_id = ? AND availability <> ?", tenantID, spu.ID, "offboarded").
			Count(&activeCount).Error; err != nil {
			return err
		}
		newStatus := spu.Status
		if activeCount == 0 {
			newStatus = "offboarded"
		}
		updatePayload := map[string]any{"updated_at": now}
		if newStatus != spu.Status {
			updatePayload["status"] = newStatus
			spu.Status = newStatus
		}
		if err := tx.Model(&spu).Updates(updatePayload).Error; err != nil {
			return err
		}
		if s.publisher != nil {
			channels := make([]string, 0, len(targetChannels))
			for _, ch := range targetChannels {
				channels = append(channels, ch.Channel)
			}
			if _, err := s.publisher.EnqueueWithdraw(ctx, tenantID, spu.ID, channels, effectiveAt, req.Reason); err != nil {
				return err
			}
		}
		detail = convertToDetail(&spu, nil, derefString(spu.CurrentVersionID))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// Delete soft-deletes an SPU when eligible.
func (s *Service) Delete(ctx context.Context, id string, req DeleteRequest) (*SPUDetail, error) {
	if s == nil {
		return nil, errors.New("spu service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		return nil, ValidationErrors{{Field: "reason", Message: "reason is required"}}
	}
	var detail *SPUDetail
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		allowed := map[string]struct{}{"draft": {}, "offboarded": {}}
		if _, ok := allowed[strings.ToLower(spu.Status)]; !ok {
			return fmt.Errorf("spu status %s cannot be deleted", spu.Status)
		}
		if err := tx.Delete(&spu).Error; err != nil {
			return err
		}
		body, _ := json.Marshal(map[string]any{
			"reason":   req.Reason,
			"operator": actorFromContext(ctx),
		})
		entry := productmodel.SPUAuditLog{
			ID:         uuidString(),
			TenantUUID: tenantID,
			SPUID:      spu.ID,
			EventType:  "spu.deleted",
			Payload:    datatypes.JSON(body),
			Operator:   actorFromContext(ctx),
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		detail = convertToDetail(&spu, nil, derefString(spu.CurrentVersionID))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// Get fetches SPU detail by id.
func (s *Service) Get(ctx context.Context, id string) (*SPUDetail, error) {
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var entity productmodel.SPU
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantID, id).
		First(&entity).Error; err != nil {
		return nil, err
	}
	return convertToDetail(&entity, nil, derefString(entity.CurrentVersionID)), nil
}

func convertToDetail(spu *productmodel.SPU, locales []LocaleContent, currentVersion string) *SPUDetail {
	var tagList []string
	if spu.Tags != nil {
		tagList = []string(spu.Tags)
	}
	return &SPUDetail{
		ID:             spu.ID,
		Code:           spu.Code,
		Name:           spu.Name,
		Type:           spu.Type,
		CategoryID:     spu.CategoryID,
		CategoryPath:   spu.CategoryPath,
		BrandID:        spu.BrandID,
		DefaultLocale:  spu.DefaultLocale,
		Status:         spu.Status,
		Tags:           tagList,
		Responsible:    spu.ResponsibleUser,
		Locales:        locales,
		CurrentVersion: currentVersion,
		CreatedAt:      spu.CreatedAt,
		UpdatedAt:      spu.UpdatedAt,
	}
}

func derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func uuidString() string {
	return uuid.NewString()
}

func pqStringArray(items []string) pq.StringArray {
	if len(items) == 0 {
		return nil
	}
	return pq.StringArray(items)
}

type approvalStage struct {
	Role     string
	Duration time.Duration
}

var defaultApprovalChain = []approvalStage{
	{Role: "ops", Duration: 24 * time.Hour},
	{Role: "qc", Duration: 24 * time.Hour},
	{Role: "legal", Duration: 48 * time.Hour},
}

func (s *Service) ensureApprovalChain(tx *gorm.DB, tenantID string, version *productmodel.SPUVersion, submitter string) error {
	if s == nil || s.approvalRepo == nil || version == nil {
		return nil
	}
	var count int64
	if err := tx.Model(&productmodel.SPUApprovalRecord{}).
		Where("tenant_uuid = ? AND version_id = ?", tenantID, version.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC()
	records := make([]productmodel.SPUApprovalRecord, 0, len(defaultApprovalChain))
	for idx, stage := range defaultApprovalChain {
		status := "pending"
		comment := ""
		var actedAt *time.Time
		actedBy := ""
		if stage.Role == "ops" && submitter != "" && submitter != "system" {
			status = "approved"
			actedBy = submitter
			actedAt = &now
			comment = "auto-approved by submitter"
		}
		due := now.Add(stage.Duration)
		record := productmodel.SPUApprovalRecord{
			ID:         uuidString(),
			TenantUUID: tenantID,
			SPUID:      version.SPUID,
			VersionID:  version.ID,
			ChainOrder: idx + 1,
			Role:       stage.Role,
			Status:     status,
			Comment:    comment,
			SLADueAt:   &due,
			ActedBy:    actedBy,
			ActedAt:    actedAt,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil
	}
	return tx.Create(&records).Error
}

func buildChannelsSummary(channels []string, state, taskID string) datatypes.JSON {
	if len(channels) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	entries := make(map[string]map[string]string, len(channels))
	for _, ch := range channels {
		channel := strings.TrimSpace(ch)
		if channel == "" {
			continue
		}
		entry := map[string]string{"status": state}
		if taskID != "" {
			entry["taskId"] = taskID
		}
		entries[channel] = entry
	}
	if len(entries) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	body, _ := json.Marshal(entries)
	return datatypes.JSON(body)
}

func dedupeStrings(items []string) []string {
	set := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		candidate := strings.TrimSpace(item)
		if candidate == "" {
			continue
		}
		key := strings.ToLower(candidate)
		if _, ok := set[key]; ok {
			continue
		}
		set[key] = struct{}{}
		result = append(result, candidate)
	}
	return result
}
