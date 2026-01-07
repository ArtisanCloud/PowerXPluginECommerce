package product_category

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	adminconsole "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/admin_console"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	consolesvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/console"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type TemplateImpactSummary struct {
	TemplateID       string   `json:"templateId"`
	BoundCategories  int64    `json:"boundCategories"`
	BoundCategoryIDs []string `json:"boundCategoryIds,omitempty"`
	AffectedProducts int64    `json:"affectedProducts"`
}

type TemplateRecheckRequest struct {
	Reason string `json:"reason,omitempty"`
	DryRun bool   `json:"dryRun,omitempty"`
}

type TemplateRecheckResult struct {
	JobRunID string `json:"jobRunId"`
	Status   string `json:"status"`
}

func (s *Service) GetTemplateImpact(ctx context.Context, templateID string) (*TemplateImpactSummary, error) {
	if s == nil || !s.Ready() || s.deps == nil || s.deps.DB == nil {
		return nil, errors.New("template impact service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil, errors.New("templateId is required")
	}
	type boundCategory struct {
		ID   string
		Path string
	}
	var categories []boundCategory
	if err := s.deps.DB.WithContext(ctx).
		Model(&productcategory.ProductCategory{}).
		Select("id, path").
		Where("tenant_uuid = ? AND template_id = ?", tenantID, templateID).
		Order("level ASC, sort_order ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	if len(categories) == 0 {
		return &TemplateImpactSummary{TemplateID: templateID, BoundCategories: 0, AffectedProducts: 0}, nil
	}
	ids := make([]string, 0, len(categories))
	var paths []string
	for _, c := range categories {
		if strings.TrimSpace(c.ID) != "" {
			ids = append(ids, strings.TrimSpace(c.ID))
		}
		if strings.TrimSpace(c.Path) != "" {
			paths = append(paths, strings.TrimSpace(c.Path))
		}
	}
	query := s.deps.DB.WithContext(ctx).Model(&productmodel.SPU{}).
		Where("tenant_uuid = ?", tenantID)
	if len(paths) > 0 {
		conds := make([]string, 0, len(paths))
		args := make([]any, 0, len(paths))
		for _, p := range paths {
			conds = append(conds, "category_path LIKE ?")
			args = append(args, p+"%")
		}
		query = query.Where("("+strings.Join(conds, " OR ")+")", args...)
	} else if len(ids) > 0 {
		query = query.Where("category_id IN ?", ids)
	}
	var productCount int64
	if err := query.Count(&productCount).Error; err != nil {
		return nil, err
	}
	return &TemplateImpactSummary{
		TemplateID:       templateID,
		BoundCategories:  int64(len(ids)),
		BoundCategoryIDs: ids,
		AffectedProducts: productCount,
	}, nil
}

func (s *Service) TriggerTemplateRecheck(ctx context.Context, templateID string, req TemplateRecheckRequest) (*TemplateRecheckResult, error) {
	if s == nil || !s.Ready() || s.deps == nil || s.deps.DB == nil {
		return nil, errors.New("template recheck service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil, errors.New("templateId is required")
	}
	jobService := consolesvc.NewJobService(s.deps)
	if jobService == nil {
		return nil, errors.New("job service unavailable")
	}

	meta := map[string]any{
		"type":       "category_template_recheck",
		"templateId": templateID,
	}
	metaBody, _ := json.Marshal(meta)
	input := consolesvc.ScheduleSafeOpInput{
		TenantUuid:    &tenantID,
		Environment:   "plugin",
		JobType:       consolesvc.JobTypeCustom,
		TriggerSource: consolesvc.TriggerSourceAPI,
		Action:        consolesvc.SafeOpAction("recheck"),
		ScopeType:     consolesvc.SafeOpScopeTenant,
		ScopeRef:      tenantID,
		TargetID:      templateID,
		Reason:        strings.TrimSpace(req.Reason),
		DryRun:        req.DryRun,
		Message:       "category template batch recheck requested",
		Actor: consolesvc.Actor{
			ID:             "system",
			PermissionCode: "admin.product.categoryTemplates.recheck",
		},
	}
	record, err := jobService.ScheduleSafeOp(ctx, input)
	if err != nil {
		return nil, err
	}

	// 额外记录一条轻量审计（可选）：写入 admin_console_audit_events 以便后续查询
	now := time.Now().UTC()
	event := &adminconsole.AuditEvent{
		ID:             utils.NewUUID(),
		PluginID:       app.PluginID,
		TenantUuid:     &tenantID,
		ActorID:        "system",
		PermissionCode: input.Actor.PermissionCode,
		Action:         "trigger_recheck",
		ResourceType:   "category_template",
		ResourceRef:    &templateID,
		Diff:           datatypes.JSON(metaBody),
		OccurredAt:     now,
		CreatedAt:      now,
	}
	_ = s.deps.DB.WithContext(ctx).Create(event).Error

	return &TemplateRecheckResult{
		JobRunID: record.ID,
		Status:   string(record.Status),
	}, nil
}
