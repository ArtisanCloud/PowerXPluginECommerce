package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

var (
	ErrCouponQueryServiceUnavailable = errors.New("coupon query service unavailable")
)

type AssetQueryFilter struct {
	TemplateID string
	UserID     string
	Status     string
	OrderID    string
	CouponCode string
	Page       int
	PageSize   int
}

type AssetListResult struct {
	Items    []couponmodel.CouponAsset `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

type UsageLogQueryFilter struct {
	AssetID    string
	TemplateID string
	OrderID    string
	Action     string
	CouponCode string
	UserID     string
	Page       int
	PageSize   int
}

type UsageLogItem struct {
	ID             string    `json:"id"`
	TenantUUID     string    `json:"tenant_uuid"`
	AssetID        string    `json:"asset_id"`
	OrderID        *string   `json:"order_id,omitempty"`
	CouponCode     string    `json:"coupon_code,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
	Action         string    `json:"action"`
	ActionReason   string    `json:"action_reason"`
	IdempotencyKey string    `json:"idempotency_key"`
	RequestID      string    `json:"request_id,omitempty"`
	CreatedBy      string    `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type UsageLogListResult struct {
	Items    []UsageLogItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type QueryService struct {
	deps *app.Deps
}

func NewQueryService(deps *app.Deps) *QueryService {
	return &QueryService{deps: deps}
}

func (s *QueryService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil
}

func (s *QueryService) ListAssets(ctx context.Context, tenantUUID string, filter AssetQueryFilter) (*AssetListResult, error) {
	if !s.Ready() {
		return nil, ErrCouponQueryServiceUnavailable
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
	query := s.deps.DB.WithContext(ctx).Model(&couponmodel.CouponAsset{}).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.TemplateID); v != "" {
		query = query.Where("template_id = ?", v)
	}
	if v := strings.TrimSpace(filter.UserID); v != "" {
		query = query.Where("user_id = ?", v)
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := strings.TrimSpace(filter.OrderID); v != "" {
		query = query.Where("reserved_order_id = ?", v)
	}
	if v := strings.TrimSpace(filter.CouponCode); v != "" {
		query = query.Where("coupon_code LIKE ?", "%"+v+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []couponmodel.CouponAsset
	if err := query.Order("updated_at DESC, created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &AssetListResult{Items: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *QueryService) ListUsageLogs(ctx context.Context, tenantUUID string, filter UsageLogQueryFilter) (*UsageLogListResult, error) {
	if !s.Ready() {
		return nil, ErrCouponQueryServiceUnavailable
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
	query := s.deps.DB.WithContext(ctx).
		Table(models.S(models.TableCouponUsageLogs)+" AS ul").
		Joins("LEFT JOIN "+models.S(models.TableCouponAssets)+" AS a ON a.tenant_uuid = ul.tenant_uuid AND a.id = ul.asset_id").
		Where("ul.tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.AssetID); v != "" {
		query = query.Where("ul.asset_id = ?", v)
	}
	if v := strings.TrimSpace(filter.TemplateID); v != "" {
		query = query.Where("a.template_id = ?", v)
	}
	if v := strings.TrimSpace(filter.OrderID); v != "" {
		query = query.Where("ul.order_id = ?", v)
	}
	if v := strings.TrimSpace(filter.Action); v != "" {
		query = query.Where("ul.action = ?", strings.ToLower(v))
	}
	if v := strings.TrimSpace(filter.CouponCode); v != "" {
		query = query.Where("a.coupon_code LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.UserID); v != "" {
		query = query.Where("a.user_id = ?", v)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	type row struct {
		ID             string
		TenantUUID     string
		AssetID        string
		OrderID        *string
		CouponCode     string
		UserID         string
		Action         string
		ActionReason   string
		IdempotencyKey string
		RequestID      string
		CreatedBy      string
		CreatedAt      time.Time
	}
	var rows []row
	if err := query.Select(`
		ul.id, ul.tenant_uuid, ul.asset_id, ul.order_id,
		COALESCE(a.coupon_code, '') AS coupon_code,
		COALESCE(a.user_id, '') AS user_id,
		ul.action, ul.action_reason, ul.idempotency_key, ul.request_id, ul.created_by, ul.created_at
	`).Order("ul.created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]UsageLogItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, UsageLogItem{
			ID:             strings.TrimSpace(r.ID),
			TenantUUID:     strings.TrimSpace(r.TenantUUID),
			AssetID:        strings.TrimSpace(r.AssetID),
			OrderID:        r.OrderID,
			CouponCode:     strings.TrimSpace(r.CouponCode),
			UserID:         strings.TrimSpace(r.UserID),
			Action:         strings.TrimSpace(r.Action),
			ActionReason:   strings.TrimSpace(r.ActionReason),
			IdempotencyKey: strings.TrimSpace(r.IdempotencyKey),
			RequestID:      strings.TrimSpace(r.RequestID),
			CreatedBy:      strings.TrimSpace(r.CreatedBy),
			CreatedAt:      r.CreatedAt,
		})
	}
	return &UsageLogListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
