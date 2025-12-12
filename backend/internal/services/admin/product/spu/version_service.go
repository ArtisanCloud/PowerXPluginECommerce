package spu

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/domain/repository/product"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// VersionService exposes helpers for version history, approvals, and rollback flows.
type VersionService struct {
	deps         *app.Deps
	spuRepo      *productrepo.SPURepository
	versionRepo  *productrepo.VersionRepository
	approvalRepo *productrepo.ApprovalRepository
}

// VersionListFilters controls pagination and filtering for version list queries.
type VersionListFilters struct {
	Status   string
	Page     int
	PageSize int
}

// VersionListResult provides pagination metadata for version listings.
type VersionListResult struct {
	Items    []VersionSummary `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// VersionSummary captures lightweight metadata for timelines.
type VersionSummary struct {
	ID            string     `json:"id"`
	VersionNumber int        `json:"versionNumber"`
	Status        string     `json:"status"`
	SubmittedBy   string     `json:"submittedBy,omitempty"`
	SubmittedAt   *time.Time `json:"submittedAt,omitempty"`
	ApprovedAt    *time.Time `json:"approvedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// DiffEntry captures a single changed field between versions.
type DiffEntry struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

// VersionApproval describes a single approval stage.
type VersionApproval struct {
	ID         string     `json:"id"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	Comment    string     `json:"comment,omitempty"`
	SLADueAt   *time.Time `json:"slaDueAt,omitempty"`
	ActedBy    string     `json:"actedBy,omitempty"`
	ActedAt    *time.Time `json:"actedAt,omitempty"`
	ChainOrder int        `json:"chainOrder"`
}

// VersionDetail aggregates payload, diff, and approval metadata.
type VersionDetail struct {
	ID               string            `json:"id"`
	SPUID            string            `json:"spuId"`
	VersionNumber    int               `json:"versionNumber"`
	Status           string            `json:"status"`
	SubmittedBy      string            `json:"submittedBy,omitempty"`
	SubmittedAt      *time.Time        `json:"submittedAt,omitempty"`
	ApprovedBy       string            `json:"approvedBy,omitempty"`
	ApprovedAt       *time.Time        `json:"approvedAt,omitempty"`
	RollbackSourceID string            `json:"rollbackSourceId,omitempty"`
	Payload          map[string]any    `json:"payload"`
	Diff             []DiffEntry       `json:"diff"`
	Approvals        []VersionApproval `json:"approvals"`
}

// ApprovalActionRequest captures approver input.
type ApprovalActionRequest struct {
	Comment string `json:"comment"`
}

// RollbackRequest describes fields required to resurrect a historical version.
type RollbackRequest struct {
	TargetVersionID string `json:"targetVersionId"`
	Reason          string `json:"reason"`
}

// NewVersionService wires dependencies for version operations.
func NewVersionService(deps *app.Deps) *VersionService {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &VersionService{
		deps:         deps,
		spuRepo:      productrepo.NewSPURepository(deps.DB),
		versionRepo:  productrepo.NewVersionRepository(deps.DB),
		approvalRepo: productrepo.NewApprovalRepository(deps.DB),
	}
}

// List returns paginated version summaries for a specific SPU.
func (s *VersionService) List(ctx context.Context, spuID string, filters VersionListFilters) (*VersionListResult, error) {
	if s == nil {
		return nil, errors.New("version service not initialized")
	}
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
	query := s.versionRepo.DB.WithContext(ctx).
		Model(&productmodel.SPUVersion{}).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var versions []productmodel.SPUVersion
	if err := query.Order("version_number DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&versions).Error; err != nil {
		return nil, err
	}
	items := make([]VersionSummary, 0, len(versions))
	for _, version := range versions {
		items = append(items, VersionSummary{
			ID:            version.ID,
			VersionNumber: version.VersionNumber,
			Status:        version.Status,
			SubmittedBy:   version.SubmittedBy,
			SubmittedAt:   version.SubmittedAt,
			ApprovedAt:    version.ApprovedAt,
			CreatedAt:     version.CreatedAt,
		})
	}
	return &VersionListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// Get returns the version payload, diff summary, and approval timeline.
func (s *VersionService) Get(ctx context.Context, spuID, versionID string) (*VersionDetail, error) {
	if s == nil {
		return nil, errors.New("version service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var spu productmodel.SPU
	if err := s.spuRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantID, spuID).
		First(&spu).Error; err != nil {
		return nil, err
	}
	var version productmodel.SPUVersion
	if err := s.versionRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, versionID, spuID).
		First(&version).Error; err != nil {
		return nil, err
	}
	payload := decodeVersionPayload(version.Payload)
	diff := s.diffAgainstBaseline(ctx, tenantID, &spu, &version, payload)
	approvals := s.fetchApprovals(ctx, tenantID, version.ID)
	return buildVersionDetail(&version, payload, diff, approvals), nil
}

// Approve advances the approval chain for a given version.
func (s *VersionService) Approve(ctx context.Context, spuID, versionID string, req ApprovalActionRequest) (*VersionDetail, error) {
	return s.handleApprovalAction(ctx, spuID, versionID, req, true)
}

// Reject marks a version as rejected and resets the SPU to draft.
func (s *VersionService) Reject(ctx context.Context, spuID, versionID string, req ApprovalActionRequest) (*VersionDetail, error) {
	return s.handleApprovalAction(ctx, spuID, versionID, req, false)
}

// Rollback clones the payload from a historical version into a new draft.
func (s *VersionService) Rollback(ctx context.Context, spuID string, req RollbackRequest) (*VersionDetail, error) {
	if s == nil {
		return nil, errors.New("version service not initialized")
	}
	if strings.TrimSpace(req.TargetVersionID) == "" {
		return nil, errors.New("targetVersionId is required")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var detail *VersionDetail
	err = s.versionRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
			return err
		}
		var target productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, req.TargetVersionID, spu.ID).First(&target).Error; err != nil {
			return err
		}
		var nextNumber int
		if err := tx.Model(&productmodel.SPUVersion{}).
			Where("tenant_uuid = ? AND spu_id = ?", tenantID, spu.ID).
			Select("COALESCE(MAX(version_number), 0)").Scan(&nextNumber).Error; err != nil {
			return err
		}
		nextNumber++
		now := time.Now().UTC()
		sourceID := target.ID
		newVersion := &productmodel.SPUVersion{
			ID:                    uuidString(),
			TenantUUID:            tenantID,
			SPUID:                 spu.ID,
			VersionNumber:         nextNumber,
			Status:                "draft",
			Payload:               target.Payload,
			DiffSummary:           datatypes.JSON([]byte("{}")),
			RollbackSourceVersion: &sourceID,
			SubmittedBy:           actorFromContext(ctx),
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		if err := tx.Create(newVersion).Error; err != nil {
			return err
		}
		if err := tx.Model(&spu).Updates(map[string]any{
			"status":             "draft",
			"current_version_id": newVersion.ID,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}
		// Cleanup stale approvals for the new draft in case of retries.
		if err := tx.Where("tenant_uuid = ? AND version_id = ?", tenantID, newVersion.ID).
			Delete(&productmodel.SPUApprovalRecord{}).Error; err != nil {
			return err
		}
		payload := decodeVersionPayload(newVersion.Payload)
		diff := s.diffAgainstBaseline(ctx, tenantID, &spu, newVersion, payload)
		approvals := []VersionApproval{}
		detail = buildVersionDetail(newVersion, payload, diff, approvals)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (s *VersionService) handleApprovalAction(ctx context.Context, spuID, versionID string, req ApprovalActionRequest, approve bool) (*VersionDetail, error) {
	if s == nil {
		return nil, errors.New("version service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var resultVersionID string
	err = s.versionRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
			return err
		}
		var version productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, versionID, spuID).First(&version).Error; err != nil {
			return err
		}
		var pending productmodel.SPUApprovalRecord
		if err := tx.Where("tenant_uuid = ? AND version_id = ? AND status = ?", tenantID, version.ID, "pending").
			Order("chain_order ASC").
			First(&pending).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		status := "rejected"
		if approve {
			status = "approved"
		}
		updates := map[string]any{
			"status":     status,
			"comment":    strings.TrimSpace(req.Comment),
			"acted_by":   actorFromContext(ctx),
			"acted_at":   now,
			"updated_at": now,
		}
		if err := tx.Model(&productmodel.SPUApprovalRecord{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, pending.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if approve {
			var remaining int64
			if err := tx.Model(&productmodel.SPUApprovalRecord{}).
				Where("tenant_uuid = ? AND version_id = ? AND status = ?", tenantID, version.ID, "pending").
				Count(&remaining).Error; err != nil {
				return err
			}
			if remaining == 0 {
				if err := tx.Model(&version).Updates(map[string]any{
					"approved_by": actorFromContext(ctx),
					"approved_at": now,
				}).Error; err != nil {
					return err
				}
			}
		} else {
			if err := tx.Model(&version).Updates(map[string]any{"status": "rejected"}).Error; err != nil {
				return err
			}
			if err := tx.Model(&spu).Updates(map[string]any{"status": "draft", "updated_at": now}).Error; err != nil {
				return err
			}
		}
		resultVersionID = version.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, spuID, resultVersionID)
}

func (s *VersionService) fetchApprovals(ctx context.Context, tenantID, versionID string) []VersionApproval {
	if s == nil || s.approvalRepo == nil {
		return nil
	}
	var records []productmodel.SPUApprovalRecord
	err := s.approvalRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND version_id = ?", tenantID, versionID).
		Order("chain_order ASC").
		Find(&records).Error
	if err != nil {
		return nil
	}
	result := make([]VersionApproval, 0, len(records))
	for _, rec := range records {
		result = append(result, VersionApproval{
			ID:         rec.ID,
			Role:       rec.Role,
			Status:     rec.Status,
			Comment:    rec.Comment,
			SLADueAt:   rec.SLADueAt,
			ActedBy:    rec.ActedBy,
			ActedAt:    rec.ActedAt,
			ChainOrder: rec.ChainOrder,
		})
	}
	return result
}

func (s *VersionService) diffAgainstBaseline(ctx context.Context, tenantID string, spu *productmodel.SPU, version *productmodel.SPUVersion, payload map[string]any) []DiffEntry {
	if spu == nil || version == nil {
		return nil
	}
	basePayload := map[string]any{}
	baselineID := derefString(spu.CurrentVersionID)
	if baselineID == "" || baselineID == version.ID {
		var published productmodel.SPUVersion
		err := s.versionRepo.DB.WithContext(ctx).
			Where("tenant_uuid = ? AND spu_id = ? AND status = ?", tenantID, spu.ID, "published").
			Order("version_number DESC").
			First(&published).Error
		if err == nil {
			baselineID = published.ID
		}
	}
	if baselineID != "" && baselineID != version.ID {
		var base productmodel.SPUVersion
		if err := s.versionRepo.DB.WithContext(ctx).
			Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, baselineID, spu.ID).
			First(&base).Error; err == nil {
			basePayload = decodeVersionPayload(base.Payload)
		}
	}
	return diffPayloadMaps(basePayload, payload)
}

func diffPayloadMaps(base, target map[string]any) []DiffEntry {
	baseFlat := make(map[string]any)
	targetFlat := make(map[string]any)
	flattenMap("", base, baseFlat)
	flattenMap("", target, targetFlat)
	keys := make([]string, 0, len(baseFlat)+len(targetFlat))
	seen := map[string]struct{}{}
	for key := range baseFlat {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range targetFlat {
		if _, ok := seen[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	diff := make([]DiffEntry, 0, len(keys))
	for _, key := range keys {
		before, beforeOK := baseFlat[key]
		after, afterOK := targetFlat[key]
		if beforeOK && afterOK && reflect.DeepEqual(before, after) {
			continue
		}
		entry := DiffEntry{Field: key}
		if beforeOK {
			entry.Before = before
		}
		if afterOK {
			entry.After = after
		}
		diff = append(diff, entry)
	}
	return diff
}

func flattenMap(prefix string, value any, out map[string]any) {
	if out == nil {
		return
	}
	switch typed := value.(type) {
	case map[string]any:
		for key, val := range typed {
			child := key
			if prefix != "" {
				child = prefix + "." + key
			}
			flattenMap(child, val, out)
		}
	case []any:
		for idx, item := range typed {
			child := fmt.Sprintf("%s[%d]", prefix, idx)
			child = strings.TrimPrefix(child, ".")
			flattenMap(child, item, out)
		}
	case nil:
		out[prefix] = nil
	default:
		key := prefix
		if key == "" {
			key = "value"
		}
		out[key] = typed
	}
}

func buildVersionDetail(entity *productmodel.SPUVersion, payload map[string]any, diff []DiffEntry, approvals []VersionApproval) *VersionDetail {
	if entity == nil {
		return nil
	}
	return &VersionDetail{
		ID:               entity.ID,
		SPUID:            entity.SPUID,
		VersionNumber:    entity.VersionNumber,
		Status:           entity.Status,
		SubmittedBy:      entity.SubmittedBy,
		SubmittedAt:      entity.SubmittedAt,
		ApprovedBy:       entity.ApprovedBy,
		ApprovedAt:       entity.ApprovedAt,
		RollbackSourceID: derefString(entity.RollbackSourceVersion),
		Payload:          payload,
		Diff:             diff,
		Approvals:        approvals,
	}
}

func (s *VersionService) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil && s.deps != nil {
		ctx = s.deps.Ctx
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return strings.TrimSpace(tid), nil
	}
	return "", ErrMissingTenant
}

func actorFromContext(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}
	if tc, ok := ctx.Value("tenant_ctx").(authx.TenantContext); ok {
		if tc.UserID > 0 {
			return fmt.Sprintf("user:%d", tc.UserID)
		}
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && tid != "" {
		return "tenant:" + tid
	}
	return "system"
}
