package spu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product"
	channelproductjobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/jobs/channel/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product"
	product_category "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ErrMissingTenant indicates the request context did not carry tenant metadata.
var ErrMissingTenant = errors.New("tenant context missing")

// ErrInvalidStatus indicates an unsupported state transition.
var ErrInvalidStatus = errors.New("invalid status transition")

var (
	ErrInvalidSPUListSort      = errors.New("invalid spu list sort")
	ErrInvalidSPUListOrder     = errors.New("invalid spu list order")
	ErrUnsupportedSPUListSort  = errors.New("unsupported spu list sort")
	ErrSPUListRequiresPostgres = errors.New("spu list sort/filter requires postgres")

	// ErrPublishRequiresInventory indicates the SPU has no saleable SKU inventory for publishing.
	ErrPublishRequiresInventory = errors.New("publish requires at least one sku with available inventory")
)

// Service orchestrates tenant-scoped SPU lifecycle operations.
type Service struct {
	deps         *app.Deps
	spuRepo      *productrepo.SPURepository
	versionRepo  *productrepo.VersionRepository
	channelRepo  *productrepo.ChannelRepository
	planRepo     *productrepo.SubscriptionPlanRepository
	approvalRepo *productrepo.ApprovalRepository
	locales      *LocaleService
	publisher    channelproductjobs.Publisher
	metrics      *productmetrics.SPUMetrics
}

// NewService wires repositories and helpers using the shared dependency bundle.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("spu service requires initialized DB dependency")
	}
	logger := deps.RuntimeLogger(context.TODO(), "product-spu-publisher", nil)
	return &Service{
		deps:         deps,
		spuRepo:      productrepo.NewSPURepository(deps.DB),
		versionRepo:  productrepo.NewVersionRepository(deps.DB),
		channelRepo:  productrepo.NewChannelRepository(deps.DB),
		planRepo:     productrepo.NewSubscriptionPlanRepository(deps.DB),
		approvalRepo: productrepo.NewApprovalRepository(deps.DB),
		locales:      NewLocaleService([]string{"zh-CN"}),
		publisher:    channelproductjobs.NewAsyncPublisher(logger),
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
	Attributes      map[string]any  `json:"attributes,omitempty"`
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
	if r.Attributes == nil {
		r.Attributes = map[string]any{}
	}
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
		// US2：按类目取生效模板并校验 attributes（仅新建/编辑校验）
		templateSvc := product_category.NewService(s.deps)
		if templateSvc != nil && templateSvc.Ready() {
			tErrs, err := templateSvc.ValidateAttributesForCategory(ctx, req.CategoryID, req.Attributes)
			if err != nil {
				return err
			}
			for _, e := range tErrs {
				errs = errs.add(e.Field, e.Message)
			}
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
	Keyword            string
	Status             string
	Type               string
	CategoryID         string
	CategoryPathPrefix string
	Tags               []string
	Sort               string
	Order              string
	MinPrice           *float64
	MaxPrice           *float64
	InStock            *bool
	HasPlans           *bool
	Page               int
	PageSize           int
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
	SKUCount  int64     `json:"skuCount"`
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
	sortKey, err := normalizeSPUListSort(filters.Sort)
	if err != nil {
		return nil, err
	}
	orderDir, err := normalizeSPUListOrder(filters.Order)
	if err != nil {
		return nil, err
	}
	if sortKey == "sales" {
		return nil, fmt.Errorf("%w: sales sort requires a sales data source", ErrUnsupportedSPUListSort)
	}
	dialect := strings.ToLower(strings.TrimSpace(s.deps.DB.Dialector.Name()))
	needsPrice := sortKey == "price" || filters.MinPrice != nil || filters.MaxPrice != nil
	needsPlans := needsPrice || filters.HasPlans != nil || filters.InStock != nil
	needsInventory := filters.InStock != nil
	if (needsPrice || needsPlans || needsInventory) && dialect != "postgres" {
		return nil, fmt.Errorf("%w: dialect=%s", ErrSPUListRequiresPostgres, dialect)
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
	if strings.TrimSpace(filters.CategoryID) != "" {
		query = query.Where("category_id = ?", strings.TrimSpace(filters.CategoryID))
	} else if strings.TrimSpace(filters.CategoryPathPrefix) != "" {
		query = query.Where("category_path LIKE ?", strings.TrimSpace(filters.CategoryPathPrefix)+"%")
	}
	if filters.Keyword != "" {
		like := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	if len(filters.Tags) > 0 {
		tagSet := map[string]struct{}{}
		tags := make([]string, 0, len(filters.Tags))
		for _, tag := range filters.Tags {
			normalized := strings.ToLower(strings.TrimSpace(tag))
			if normalized == "" {
				continue
			}
			if _, exists := tagSet[normalized]; exists {
				continue
			}
			tagSet[normalized] = struct{}{}
			tags = append(tags, normalized)
		}
		if len(tags) > 0 {
			if dialect == "postgres" {
				query = query.Where("tags && ?", pq.Array(tags))
			} else {
				conditions := make([]string, 0, len(tags))
				args := make([]any, 0, len(tags))
				for _, tag := range tags {
					conditions = append(conditions, "LOWER(tags) LIKE ?")
					args = append(args, "%"+tag+"%")
				}
				query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
			}
		}
	}

	if dialect == "postgres" {
		if needsPrice {
			skuAgg := s.deps.DB.WithContext(ctx).
				Table(productskumodel.ProductSKU{}.TableName()).
				Select("spu_id, MIN("+postgresSKUPriceExpr()+") AS min_price, MAX("+postgresSKUPriceExpr()+") AS max_price").
				Where("tenant_uuid = ? AND deleted_at IS NULL AND status = ?", tenantID, "published").
				Group("spu_id")
			query = query.Joins("LEFT JOIN (?) AS sku_agg ON sku_agg.spu_id = product_spus.id", skuAgg)
		}
		if needsPlans {
			planAgg := s.deps.DB.WithContext(ctx).
				Table(productmodel.SubscriptionPlan{}.TableName()).
				Select("spu_id, MIN(price) AS min_price, MAX(price) AS max_price, COUNT(*) AS plan_count").
				Where("tenant_uuid = ? AND status = ?", tenantID, "active").
				Group("spu_id")
			query = query.Joins("LEFT JOIN (?) AS plan_agg ON plan_agg.spu_id = product_spus.id", planAgg)
		}
		if needsInventory {
			invAgg := s.deps.DB.WithContext(ctx).
				Table(fmt.Sprintf("%s AS s", productskumodel.ProductSKU{}.TableName())).
				Select("s.spu_id, SUM(GREATEST(i.available_qty - i.locked_qty, 0)) AS available_qty").
				Joins("JOIN "+fmt.Sprintf("%s AS i", productskumodel.ProductSKUInventory{}.TableName())+" ON i.sku_id = s.id").
				Where("s.tenant_uuid = ? AND s.deleted_at IS NULL AND s.status = ? AND i.tenant_uuid = ? AND i.deleted_at IS NULL",
					tenantID, "published", tenantID).
				Group("s.spu_id")
			query = query.Joins("LEFT JOIN (?) AS inv_agg ON inv_agg.spu_id = product_spus.id", invAgg)
		}

		if filters.HasPlans != nil {
			if *filters.HasPlans {
				query = query.Where("product_spus.type = ? AND plan_agg.plan_count IS NOT NULL AND plan_agg.plan_count > 0", "subscription")
			} else {
				query = query.Where("product_spus.type <> ? OR plan_agg.plan_count IS NULL OR plan_agg.plan_count = 0", "subscription")
			}
		}
		if filters.InStock != nil {
			if *filters.InStock {
				query = query.Where(`(
product_spus.type = 'subscription' AND plan_agg.plan_count IS NOT NULL AND plan_agg.plan_count > 0
) OR (
product_spus.type <> 'subscription' AND COALESCE(inv_agg.available_qty, 0) > 0
)`)
			} else {
				query = query.Where(`(
product_spus.type = 'subscription' AND (plan_agg.plan_count IS NULL OR plan_agg.plan_count = 0)
) OR (
product_spus.type <> 'subscription' AND COALESCE(inv_agg.available_qty, 0) <= 0
)`)
			}
		}
		if filters.MinPrice != nil {
			query = query.Where(postgresSPUEffectiveMinPriceExpr()+" >= ?", *filters.MinPrice)
		}
		if filters.MaxPrice != nil {
			query = query.Where(postgresSPUEffectiveMinPriceExpr()+" <= ?", *filters.MaxPrice)
		}
	}

	// SKU count is always useful for UI list rendering (independent of pricing/inventory joins).
	skuCountAgg := s.deps.DB.WithContext(ctx).
		Table(productskumodel.ProductSKU{}.TableName()).
		Select("spu_id, COUNT(*) AS sku_count").
		Where("tenant_uuid = ? AND deleted_at IS NULL", tenantID).
		Group("spu_id")
	query = query.Joins("LEFT JOIN (?) AS sku_cnt ON sku_cnt.spu_id = product_spus.id", skuCountAgg)

	countQuery := query.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	listQuery := query.Session(&gorm.Session{}).
		Select("product_spus.id, product_spus.code, product_spus.name, product_spus.type, product_spus.status, product_spus.updated_at, COALESCE(sku_cnt.sku_count, 0) AS sku_count")
	switch sortKey {
	case "price":
		listQuery = listQuery.Order(postgresSPUEffectiveMinPriceExpr() + " " + orderDir + " NULLS LAST").Order("updated_at DESC")
	default:
		listQuery = listQuery.Order("updated_at " + orderDir)
	}

	type spuListRow struct {
		ID        string    `gorm:"column:id"`
		Code      string    `gorm:"column:code"`
		Name      string    `gorm:"column:name"`
		Type      string    `gorm:"column:type"`
		Status    string    `gorm:"column:status"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
		SKUCount  int64     `gorm:"column:sku_count"`
	}

	var records []spuListRow
	if err := listQuery.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&records).Error; err != nil {
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
			SKUCount:  rec.SKUCount,
			UpdatedAt: rec.UpdatedAt,
		})
	}
	return &ListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func normalizeSPUListSort(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "comprehensive") {
		return "updated_at", nil
	}
	switch strings.ToLower(raw) {
	case "updatedat", "updated_at":
		return "updated_at", nil
	case "price":
		return "price", nil
	case "sales":
		return "sales", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidSPUListSort, raw)
	}
}

func normalizeSPUListOrder(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "DESC", nil
	}
	switch strings.ToLower(raw) {
	case "asc":
		return "ASC", nil
	case "desc":
		return "DESC", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidSPUListOrder, raw)
	}
}

func postgresSKUPriceExpr() string {
	return `COALESCE(
NULLIF(default_values->>'sale_price','')::numeric,
NULLIF(default_values->>'salePrice','')::numeric,
NULLIF(default_values->>'price','')::numeric,
NULLIF(default_values->>'list_price','')::numeric,
NULLIF(default_values->>'listPrice','')::numeric
)`
}

func postgresSPUEffectiveMinPriceExpr() string {
	return `CASE WHEN product_spus.type = 'subscription' THEN plan_agg.min_price ELSE sku_agg.min_price END`
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
			ID:              utils.NewUUID(),
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
			ID:            utils.NewUUID(),
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
			versionID = utils.NewUUID()
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

// ReviseRequest captures audit info when creating a new draft revision.
type ReviseRequest struct {
	Reason string `json:"reason"`
}

// DeleteRequest captures audit info when removing an SPU.
type DeleteRequest struct {
	Reason string `json:"reason"`
}

// Revise creates a new draft version from a published SPU and moves it back to draft state.
// 说明：当前实现采用“单状态”模型 —— 一旦创建草稿，SPU 状态会从 published 切回 draft，但已发布版本仍保留在版本历史中。
func (s *Service) Revise(ctx context.Context, id string, req ReviseRequest) (*SPUDetail, error) {
	if s == nil {
		return nil, errors.New("spu service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		var vErrs ValidationErrors
		vErrs = vErrs.add("reason", "reason is required")
		return nil, vErrs
	}
	operator := actorFromContext(ctx)
	var detail *SPUDetail
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, id).First(&spu).Error; err != nil {
			return err
		}
		if strings.ToLower(spu.Status) != "published" {
			return fmt.Errorf("spu status %s cannot revise", spu.Status)
		}
		sourceVersionID := derefString(spu.CurrentVersionID)
		if strings.TrimSpace(sourceVersionID) == "" {
			return errors.New("missing published version")
		}
		var sourceVersion productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, sourceVersionID, spu.ID).First(&sourceVersion).Error; err != nil {
			return err
		}

		var nextVersionNumber int
		if err := tx.Model(&productmodel.SPUVersion{}).
			Where("tenant_uuid = ? AND spu_id = ?", tenantID, spu.ID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&nextVersionNumber).Error; err != nil {
			return err
		}
		nextVersionNumber += 1

		payload := decodeVersionPayload(sourceVersion.Payload)
		payload["version"] = "draft"
		payload["revisionSourceVersionId"] = sourceVersionID
		payload["revisionReason"] = req.Reason
		if operator != "" && operator != "system" {
			payload["revisionBy"] = operator
		}
		body, _ := json.Marshal(payload)

		now := time.Now().UTC()
		version := &productmodel.SPUVersion{
			ID:            utils.NewUUID(),
			TenantUUID:    tenantID,
			SPUID:         spu.ID,
			VersionNumber: nextVersionNumber,
			Status:        "draft",
			Payload:       datatypes.JSON(body),
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		if err := tx.Model(&spu).Updates(map[string]any{
			"status":             "draft",
			"current_version_id": version.ID,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}
		spu.Status = "draft"
		spu.UpdatedAt = now
		spu.CurrentVersionID = &version.ID
		detail = convertToDetail(&spu, nil, version.ID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return detail, nil
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

		ok, err := hasSaleableInventoryForSPUTx(tx, tenantID, spu.ID, "default")
		if err != nil {
			return err
		}
		if !ok {
			return ErrPublishRequiresInventory
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

func hasSaleableInventoryForSPUTx(tx *gorm.DB, tenantID, spuID, warehouseID string) (bool, error) {
	if tx == nil {
		return false, errors.New("transaction is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	spuID = strings.TrimSpace(spuID)
	warehouseID = strings.TrimSpace(warehouseID)
	if tenantID == "" || spuID == "" || warehouseID == "" {
		return false, errors.New("tenant/spu/warehouse is required")
	}

	var hit int
	err := tx.
		Table(productskumodel.ProductSKU{}.TableName() + " AS s").
		Select("1").
		Joins(
			"JOIN "+productskumodel.ProductSKUInventory{}.TableName()+" AS i ON "+
				"i.tenant_uuid = s.tenant_uuid AND i.sku_id = s.id AND i.deleted_at IS NULL",
		).
		Where(
			"s.tenant_uuid = ? AND s.deleted_at IS NULL AND s.spu_id = ? AND i.warehouse_id = ? AND i.available_qty > 0",
			tenantID, spuID, warehouseID,
		).
		Limit(1).
		Scan(&hit).Error
	if err != nil {
		return false, err
	}
	return hit == 1, nil
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
			ID:         utils.NewUUID(),
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
			ID:         utils.NewUUID(),
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
