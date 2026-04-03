package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrCouponTemplateServiceUnavailable = errors.New("coupon template service unavailable")
	ErrCouponTemplateNotFound           = errors.New("coupon template not found")
)

type TemplateListFilter struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type TemplateListResult struct {
	Items    []couponmodel.CouponTemplate `json:"items"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}

type TemplateCreateInput struct {
	TenantUUID    string         `json:"tenant_uuid"`
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	CouponType    string         `json:"coupon_type"`
	ThresholdRule map[string]any `json:"threshold_rule,omitempty"`
	ScopeRule     map[string]any `json:"scope_rule,omitempty"`
	StackingRule  map[string]any `json:"stacking_rule,omitempty"`
	RefundRule    map[string]any `json:"refund_rule,omitempty"`
	ValidFrom     time.Time      `json:"valid_from"`
	ValidTo       time.Time      `json:"valid_to"`
	Status        string         `json:"status"`
}

type TemplateUpdateInput struct {
	Name          *string         `json:"name,omitempty"`
	Status        *string         `json:"status,omitempty"`
	CouponType    *string         `json:"coupon_type,omitempty"`
	ThresholdRule *map[string]any `json:"threshold_rule,omitempty"`
	ScopeRule     *map[string]any `json:"scope_rule,omitempty"`
	StackingRule  *map[string]any `json:"stacking_rule,omitempty"`
	RefundRule    *map[string]any `json:"refund_rule,omitempty"`
	ValidFrom     *time.Time      `json:"valid_from,omitempty"`
	ValidTo       *time.Time      `json:"valid_to,omitempty"`
}

type TemplateService struct {
	deps *app.Deps
	repo *couponrepo.TemplateRepository
}

func NewTemplateService(deps *app.Deps) *TemplateService {
	if deps == nil || deps.DB == nil {
		return &TemplateService{deps: deps}
	}
	return &TemplateService{
		deps: deps,
		repo: couponrepo.NewTemplateRepository(deps.DB),
	}
}

func (s *TemplateService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *TemplateService) List(ctx context.Context, tenantUUID string, filter TemplateListFilter) (*TemplateListResult, error) {
	if !s.Ready() {
		return nil, ErrCouponTemplateServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant uuid is required")
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	query := s.deps.DB.WithContext(ctx).Model(&couponmodel.CouponTemplate{}).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.Keyword); v != "" {
		like := "%" + v + "%"
		query = query.Where("(code LIKE ? OR name LIKE ?)", like, like)
	}
	if v := strings.TrimSpace(strings.ToLower(filter.Status)); v != "" {
		query = query.Where("status = ?", v)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []couponmodel.CouponTemplate
	if err := query.Order("updated_at DESC, created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &TemplateListResult{Items: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *TemplateService) Create(ctx context.Context, in TemplateCreateInput) (*couponmodel.CouponTemplate, error) {
	if !s.Ready() {
		return nil, ErrCouponTemplateServiceUnavailable
	}
	row, err := buildTemplateModel(in)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *TemplateService) Update(ctx context.Context, tenantUUID, id string, in TemplateUpdateInput) (*couponmodel.CouponTemplate, error) {
	if !s.Ready() {
		return nil, ErrCouponTemplateServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" || id == "" {
		return nil, errors.New("tenant uuid and id are required")
	}

	var row couponmodel.CouponTemplate
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponTemplateNotFound
		}
		return nil, err
	}

	updates := map[string]any{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errors.New("name is required")
		}
		updates["name"] = name
	}
	if in.Status != nil {
		status, err := normalizeTemplateStatus(*in.Status)
		if err != nil {
			return nil, err
		}
		updates["status"] = status
	}
	if in.CouponType != nil {
		couponType := strings.TrimSpace(*in.CouponType)
		if couponType == "" {
			return nil, errors.New("coupon type is required")
		}
		updates["coupon_type"] = couponType
	}
	if in.ValidFrom != nil {
		updates["valid_from"] = in.ValidFrom.UTC()
	}
	if in.ValidTo != nil {
		updates["valid_to"] = in.ValidTo.UTC()
	}
	var nextValidFrom time.Time
	var nextValidTo time.Time
	if in.ValidFrom != nil {
		nextValidFrom = in.ValidFrom.UTC()
	} else {
		nextValidFrom = row.ValidFrom
	}
	if in.ValidTo != nil {
		nextValidTo = in.ValidTo.UTC()
	} else {
		nextValidTo = row.ValidTo
	}
	if nextValidFrom.After(nextValidTo) {
		return nil, errors.New("valid_from must be before or equal to valid_to")
	}
	if in.ThresholdRule != nil {
		raw, err := mapToJSON(*in.ThresholdRule)
		if err != nil {
			return nil, err
		}
		updates["threshold_rule"] = raw
	}
	if in.ScopeRule != nil {
		raw, err := mapToJSON(*in.ScopeRule)
		if err != nil {
			return nil, err
		}
		updates["scope_rule"] = raw
	}
	if in.StackingRule != nil {
		raw, err := mapToJSON(*in.StackingRule)
		if err != nil {
			return nil, err
		}
		updates["stacking_rule"] = raw
	}
	if in.RefundRule != nil {
		raw, err := mapToJSON(*in.RefundRule)
		if err != nil {
			return nil, err
		}
		updates["refund_rule"] = raw
	}
	if err := s.deps.DB.WithContext(ctx).Session(&gorm.Session{SkipHooks: true}).
		Model(&couponmodel.CouponTemplate{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func buildTemplateModel(in TemplateCreateInput) (*couponmodel.CouponTemplate, error) {
	in.TenantUUID = strings.TrimSpace(in.TenantUUID)
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	in.CouponType = strings.TrimSpace(in.CouponType)
	if in.TenantUUID == "" || in.Code == "" || in.Name == "" || in.CouponType == "" {
		return nil, errors.New("tenant uuid, code, name and coupon type are required")
	}
	if in.ValidFrom.IsZero() || in.ValidTo.IsZero() {
		return nil, errors.New("valid_from and valid_to are required")
	}
	if in.ValidFrom.After(in.ValidTo) {
		return nil, errors.New("valid_from must be before or equal to valid_to")
	}
	status, err := normalizeTemplateStatus(in.Status)
	if err != nil {
		return nil, err
	}
	thresholdRule, err := mapToJSON(in.ThresholdRule)
	if err != nil {
		return nil, err
	}
	scopeRule, err := mapToJSON(in.ScopeRule)
	if err != nil {
		return nil, err
	}
	stackingRule, err := mapToJSON(in.StackingRule)
	if err != nil {
		return nil, err
	}
	refundRule, err := mapToJSON(in.RefundRule)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &couponmodel.CouponTemplate{
		ID:            uuid.NewString(),
		TenantUUID:    in.TenantUUID,
		Code:          in.Code,
		Name:          in.Name,
		CouponType:    in.CouponType,
		ThresholdRule: thresholdRule,
		ScopeRule:     scopeRule,
		StackingRule:  stackingRule,
		RefundRule:    refundRule,
		ValidFrom:     in.ValidFrom.UTC(),
		ValidTo:       in.ValidTo.UTC(),
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func normalizeTemplateStatus(raw string) (string, error) {
	status := strings.TrimSpace(strings.ToLower(raw))
	if status == "" {
		return "draft", nil
	}
	switch status {
	case "draft", "active", "inactive", "disabled", "archived":
		return status, nil
	default:
		return "", errors.New("invalid template status")
	}
}

func mapToJSON(payload map[string]any) (datatypes.JSON, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	return datatypes.JSONMap(payload).MarshalJSON()
}
