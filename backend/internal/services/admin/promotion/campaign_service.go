package promotion

import (
	"context"
	"errors"
	"strings"
	"time"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPromotionServiceUnavailable = errors.New("promotion service unavailable")
	ErrPromotionNotFound           = errors.New("promotion not found")
)

type CampaignListFilter struct {
	Keyword       string
	PromotionType string
	Status        string
	Channel       string
	Page          int
	PageSize      int
}

type CampaignListResult struct {
	Items    []promotionmodel.Campaign `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

type CampaignCreateInput struct {
	TenantUUID    string
	Code          string
	Name          string
	Description   string
	PromotionType string
	ConditionRule ConditionRule
	ScopeRule     ScopeRule
	ActionRule    ActionRule
	StackingRule  StackingRule
	ValidFrom     time.Time
	ValidTo       time.Time
	SaveAction    string
	Actor         string
	RequestID     string
}

type CampaignUpdateInput struct {
	Name          *string
	Description   *string
	ConditionRule *ConditionRule
	ScopeRule     *ScopeRule
	ActionRule    *ActionRule
	StackingRule  *StackingRule
	ValidFrom     *time.Time
	ValidTo       *time.Time
	Actor         string
	RequestID     string
}

type CampaignService struct {
	deps  *app.Deps
	repo  *promotionrepo.CampaignRepository
	audit *AuditLogService
}

func NewCampaignService(deps *app.Deps) *CampaignService {
	if deps == nil || deps.DB == nil {
		return &CampaignService{deps: deps}
	}
	return &CampaignService{deps: deps, repo: promotionrepo.NewCampaignRepository(deps.DB), audit: NewAuditLogService(deps)}
}

func (s *CampaignService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *CampaignService) List(ctx context.Context, tenantUUID string, filter CampaignListFilter) (*CampaignListResult, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	query := s.deps.DB.WithContext(ctx).Model(&promotionmodel.Campaign{}).Where("tenant_uuid = ?", strings.TrimSpace(tenantUUID))
	if v := strings.TrimSpace(filter.Keyword); v != "" {
		like := "%" + v + "%"
		query = query.Where("(code LIKE ? OR name LIKE ?)", like, like)
	}
	if v := strings.TrimSpace(filter.PromotionType); v != "" {
		query = query.Where("promotion_type = ?", v)
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []promotionmodel.Campaign
	if err := query.Order("updated_at DESC, created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &CampaignListResult{Items: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *CampaignService) Create(ctx context.Context, in CampaignCreateInput) (*promotionmodel.Campaign, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	status := promotionmodel.StatusDraft
	if strings.EqualFold(strings.TrimSpace(in.SaveAction), "activate") {
		status = promotionmodel.StatusActive
	}
	if err := ValidateCampaignRules(in.PromotionType, in.ValidFrom, in.ValidTo, in.ConditionRule, in.ScopeRule, in.ActionRule, in.StackingRule); err != nil {
		return nil, err
	}
	row, err := campaignFromInput(in, status)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, nil, row.TenantUUID, row.ID, "", promotionmodel.AuditActionCreate, "", in.Actor, in.RequestID, row)
	if status == promotionmodel.StatusActive {
		_ = s.audit.Log(ctx, nil, row.TenantUUID, row.ID, "", promotionmodel.AuditActionActivate, "", in.Actor, in.RequestID, nil)
	}
	return row, nil
}

func (s *CampaignService) Update(ctx context.Context, tenantUUID, id string, in CampaignUpdateInput) (*promotionmodel.Campaign, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	row, err := s.repo.GetByID(ctx, tenantUUID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPromotionNotFound
		}
		return nil, err
	}
	condition := ValueFromJSON(row.ConditionRule, ConditionRule{})
	scope := ValueFromJSON(row.ScopeRule, ScopeRule{ScopeType: "all"})
	action := ValueFromJSON(row.ActionRule, ActionRule{})
	stacking := ValueFromJSON(row.StackingRule, StackingRule{Priority: 100, Stackable: true, StackableWithCoupon: true})
	validFrom, validTo := row.ValidFrom, row.ValidTo
	if in.ConditionRule != nil {
		condition = *in.ConditionRule
	}
	if in.ScopeRule != nil {
		scope = *in.ScopeRule
	}
	if in.ActionRule != nil {
		action = *in.ActionRule
	}
	if in.StackingRule != nil {
		stacking = *in.StackingRule
	}
	if in.ValidFrom != nil {
		validFrom = in.ValidFrom.UTC()
	}
	if in.ValidTo != nil {
		validTo = in.ValidTo.UTC()
	}
	if err := ValidateCampaignRules(row.PromotionType, validFrom, validTo, condition, scope, action, stacking); err != nil {
		return nil, err
	}
	updates := map[string]any{"updated_at": time.Now().UTC(), "updated_by": strings.TrimSpace(in.Actor), "valid_from": validFrom, "valid_to": validTo}
	if in.Name != nil {
		updates["name"] = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		updates["description"] = strings.TrimSpace(*in.Description)
	}
	if raw, err := JSONFromValue(condition); err == nil {
		updates["condition_rule"] = raw
	} else {
		return nil, err
	}
	if raw, err := JSONFromValue(scope); err == nil {
		updates["scope_rule"] = raw
	} else {
		return nil, err
	}
	if raw, err := JSONFromValue(action); err == nil {
		updates["action_rule"] = raw
	} else {
		return nil, err
	}
	if raw, err := JSONFromValue(stacking); err == nil {
		updates["stacking_rule"] = raw
	} else {
		return nil, err
	}
	if err := s.deps.DB.WithContext(ctx).Model(&promotionmodel.Campaign{}).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, nil, tenantUUID, id, "", promotionmodel.AuditActionUpdate, "", in.Actor, in.RequestID, updates)
	return s.repo.GetByID(ctx, tenantUUID, id)
}

func (s *CampaignService) Activate(ctx context.Context, tenantUUID, id, actor, requestID string) (*promotionmodel.Campaign, error) {
	return s.changeStatus(ctx, tenantUUID, id, promotionmodel.StatusActive, promotionmodel.AuditActionActivate, actor, requestID)
}

func (s *CampaignService) Pause(ctx context.Context, tenantUUID, id, actor, requestID string) (*promotionmodel.Campaign, error) {
	return s.changeStatus(ctx, tenantUUID, id, promotionmodel.StatusPaused, promotionmodel.AuditActionPause, actor, requestID)
}

func (s *CampaignService) Clone(ctx context.Context, tenantUUID, id, actor, requestID string) (*promotionmodel.Campaign, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	src, err := s.repo.GetByID(ctx, tenantUUID, id)
	if err != nil {
		return nil, err
	}
	dst := *src
	dst.ID = uuid.NewString()
	dst.Code = src.Code + "_COPY_" + time.Now().UTC().Format("150405")
	dst.Name = src.Name + " Copy"
	dst.Status = promotionmodel.StatusDraft
	dst.CreatedBy = actor
	dst.UpdatedBy = actor
	dst.CreatedAt = time.Time{}
	dst.UpdatedAt = time.Time{}
	dst.DeletedAt = gorm.DeletedAt{}
	if _, err := s.repo.Create(ctx, &dst); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, nil, tenantUUID, dst.ID, "", promotionmodel.AuditActionClone, "", actor, requestID, map[string]any{"source_id": id})
	return &dst, nil
}

func (s *CampaignService) changeStatus(ctx context.Context, tenantUUID, id, status, action, actor, requestID string) (*promotionmodel.Campaign, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	row, err := s.repo.GetByID(ctx, tenantUUID, id)
	if err != nil {
		return nil, err
	}
	condition := ValueFromJSON(row.ConditionRule, ConditionRule{})
	scope := ValueFromJSON(row.ScopeRule, ScopeRule{ScopeType: "all"})
	ruleAction := ValueFromJSON(row.ActionRule, ActionRule{})
	stacking := ValueFromJSON(row.StackingRule, StackingRule{Priority: 100, Stackable: true, StackableWithCoupon: true})
	if status == promotionmodel.StatusActive {
		if err := ValidateCampaignRules(row.PromotionType, row.ValidFrom, row.ValidTo, condition, scope, ruleAction, stacking); err != nil {
			return nil, err
		}
	}
	if err := s.deps.DB.WithContext(ctx).Model(&promotionmodel.Campaign{}).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).Updates(map[string]any{
		"status": status, "updated_by": strings.TrimSpace(actor), "updated_at": time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, nil, tenantUUID, id, "", action, "", actor, requestID, nil)
	return s.repo.GetByID(ctx, tenantUUID, id)
}

func campaignFromInput(in CampaignCreateInput, status string) (*promotionmodel.Campaign, error) {
	condition, err := JSONFromValue(in.ConditionRule)
	if err != nil {
		return nil, err
	}
	scope, err := JSONFromValue(in.ScopeRule)
	if err != nil {
		return nil, err
	}
	action, err := JSONFromValue(in.ActionRule)
	if err != nil {
		return nil, err
	}
	stacking, err := JSONFromValue(in.StackingRule)
	if err != nil {
		return nil, err
	}
	return &promotionmodel.Campaign{
		TenantUUID: strings.TrimSpace(in.TenantUUID), Code: strings.TrimSpace(in.Code), Name: strings.TrimSpace(in.Name),
		Description: strings.TrimSpace(in.Description), PromotionType: strings.TrimSpace(in.PromotionType),
		ConditionRule: condition, ScopeRule: scope, ActionRule: action, StackingRule: stacking,
		ValidFrom: in.ValidFrom.UTC(), ValidTo: in.ValidTo.UTC(), Status: status,
		CreatedBy: strings.TrimSpace(in.Actor), UpdatedBy: strings.TrimSpace(in.Actor),
	}, nil
}
