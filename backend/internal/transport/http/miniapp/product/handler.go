package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	pricingmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	skuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Handler exposes read-only product/SKU endpoints to mini-app clients.
type Handler struct {
	spuService *spuservice.Service
	skuService *skuservice.Service
	planSvc    *spuservice.SubscriptionPlanService
	db         *gorm.DB
}

var miniAppVisibleSKUStatuses = []string{"online", "ready", "published"}

func (h *Handler) beginTenantTx(ctx context.Context, tenantUUID string) (*gorm.DB, func()) {
	if h == nil || h.db == nil {
		return nil, func() {}
	}
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, func() {}
	}
	cleanup := func() { _ = tx.Rollback() }
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		if err := tx.Exec("SELECT set_config('app.tenant_uuid', ?, true)", strings.TrimSpace(tenantUUID)).Error; err != nil {
			cleanup()
			return nil, func() {}
		}
	}
	return tx, cleanup
}

func clampInt64ToIntPtr(v int64) *int {
	if v < 0 {
		v = 0
	}
	maxInt := int64(^uint(0) >> 1)
	if v > maxInt {
		v = maxInt
	}
	n := int(v)
	return &n
}

// NewHandler wires SPU/SKU services for handlers.
func NewHandler(spuSvc *spuservice.Service, skuSvc *skuservice.Service, planSvc *spuservice.SubscriptionPlanService, db *gorm.DB) *Handler {
	return &Handler{
		spuService: spuSvc,
		skuService: skuSvc,
		planSvc:    planSvc,
		db:         db,
	}
}

// ListProducts returns a trimmed SPU list for mobile clients.
func (h *Handler) ListProducts(c *gin.Context) {
	if h == nil || h.spuService == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("product service unavailable"))
		return
	}
	var query miniAppProductListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondMiniAppError(c, http.StatusBadRequest, err)
		return
	}
	status := strings.TrimSpace(query.Status)
	if status == "" {
		status = "published"
	}
	if !strings.EqualFold(status, "published") {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("only published products are accessible via mini-app"))
		return
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Q)
	}
	tags := parseTagQuery(query.Tags, query.Tag)
	result, err := h.spuService.List(c.Request.Context(), spuservice.ListFilters{
		Keyword:            keyword,
		Status:             status,
		Type:               strings.TrimSpace(query.Type),
		CategoryID:         strings.TrimSpace(query.CategoryID),
		CategoryPathPrefix: strings.TrimSpace(query.CategoryPathPrefix),
		Tags:               tags,
		Sort:               strings.TrimSpace(query.Sort),
		Order:              strings.TrimSpace(query.Order),
		MinPrice:           query.MinPrice,
		MaxPrice:           query.MaxPrice,
		InStock:            query.InStock,
		HasPlans:           query.HasPlans,
		Page:               query.Page,
		PageSize:           query.PageSize,
	})
	if err != nil {
		if errors.Is(err, spuservice.ErrInvalidSPUListSort) ||
			errors.Is(err, spuservice.ErrInvalidSPUListOrder) ||
			errors.Is(err, spuservice.ErrUnsupportedSPUListSort) ||
			errors.Is(err, spuservice.ErrSPUListRequiresPostgres) {
			respondMiniAppError(c, http.StatusBadRequest, err)
			return
		}
		respondMiniAppError(c, http.StatusInternalServerError, err)
		return
	}
	if result == nil {
		result = &spuservice.ListResult{
			Items:    []spuservice.SPUSummary{},
			Page:     query.Page,
			PageSize: query.PageSize,
		}
	}

	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	enriched := h.enrichProductSummaries(c.Request.Context(), tenantUUID, result.Items)

	resp := productListResponse{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Items:    make([]miniAppProductSummary, 0, len(enriched)),
	}
	resp.Items = append(resp.Items, enriched...)
	contracts.ResponseSuccess(c, resp)
}

// ListProductTags returns the distinct tag list (and product count) for published SPUs.
func (h *Handler) ListProductTags(c *gin.Context) {
	if h == nil || h.db == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("product service unavailable"))
		return
	}
	var query miniAppProductTagListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondMiniAppError(c, http.StatusBadRequest, err)
		return
	}
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	if strings.TrimSpace(tenantUUID) == "" {
		respondMiniAppError(c, http.StatusUnauthorized, errors.New("tenant context missing"))
		return
	}

	categoryID := strings.TrimSpace(query.CategoryID)
	categoryPathPrefix := strings.TrimSpace(query.CategoryPathPrefix)

	dialect := strings.ToLower(strings.TrimSpace(h.db.Dialector.Name()))
	items := make([]productTagItem, 0)
	if dialect == "postgres" {
		type row struct {
			Tag   string `gorm:"column:tag"`
			Count int    `gorm:"column:count"`
		}
		var rows []row
		spuTable := productmodel.SPU{}.TableName()
		sql := fmt.Sprintf(`SELECT LOWER(tag) AS tag, COUNT(*) AS count
FROM %s AS spus, UNNEST(spus.tags) AS tag
WHERE spus.tenant_uuid = ? AND spus.deleted_at IS NULL AND spus.status = ?`, spuTable)
		args := []any{tenantUUID, "published"}
		if categoryID != "" {
			sql += " AND spus.category_id = ?"
			args = append(args, categoryID)
		} else if categoryPathPrefix != "" {
			sql += " AND spus.category_path LIKE ?"
			args = append(args, categoryPathPrefix+"%")
		}
		sql += " GROUP BY LOWER(tag) ORDER BY count DESC, tag ASC LIMIT ?"
		args = append(args, limit)
		if err := h.db.WithContext(c.Request.Context()).Raw(sql, args...).Scan(&rows).Error; err != nil {
			respondMiniAppError(c, http.StatusInternalServerError, err)
			return
		}
		for _, r := range rows {
			tag := strings.TrimSpace(r.Tag)
			if tag == "" {
				continue
			}
			items = append(items, productTagItem{Tag: tag, Count: r.Count})
		}
		contracts.ResponseSuccess(c, productTagListResponse{Items: items})
		return
	}

	type tagCell struct {
		Tags string `gorm:"column:tags"`
	}
	var rows []tagCell
	tx := h.db.WithContext(c.Request.Context()).
		Table(productmodel.SPU{}.TableName()).
		Select("tags").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND status = ?", tenantUUID, "published")
	if categoryID != "" {
		tx = tx.Where("category_id = ?", categoryID)
	} else if categoryPathPrefix != "" {
		tx = tx.Where("category_path LIKE ?", categoryPathPrefix+"%")
	}
	if err := tx.Find(&rows).Error; err != nil {
		respondMiniAppError(c, http.StatusInternalServerError, err)
		return
	}
	counter := map[string]int{}
	for _, r := range rows {
		tags := parseSPUTagsCell(r.Tags)
		if len(tags) == 0 {
			continue
		}
		perRow := map[string]struct{}{}
		for _, tag := range tags {
			if tag == "" {
				continue
			}
			tag = strings.ToLower(strings.TrimSpace(tag))
			if tag == "" {
				continue
			}
			if _, exists := perRow[tag]; exists {
				continue
			}
			perRow[tag] = struct{}{}
			counter[tag]++
		}
	}
	for tag, count := range counter {
		items = append(items, productTagItem{Tag: tag, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Tag < items[j].Tag
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > limit {
		items = items[:limit]
	}
	contracts.ResponseSuccess(c, productTagListResponse{Items: items})
}

func parseTagQuery(tagsRaw, tagRaw string) []string {
	combined := strings.TrimSpace(tagsRaw)
	if combined == "" {
		combined = strings.TrimSpace(tagRaw)
	}
	if combined == "" {
		return nil
	}
	parts := strings.Split(combined, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		tag := strings.ToLower(strings.TrimSpace(part))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseSPUTagsCell(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var tags []string
		if err := json.Unmarshal([]byte(raw), &tags); err == nil {
			return tags
		}
	}
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "{"), "}"))
		if body == "" {
			return nil
		}
		parts := strings.Split(body, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			tag := strings.TrimSpace(p)
			tag = strings.Trim(tag, `"`)
			if tag != "" {
				out = append(out, tag)
			}
		}
		return out
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		tag := strings.TrimSpace(p)
		if tag != "" {
			out = append(out, tag)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// GetProduct returns a single published SPU detail.
func (h *Handler) GetProduct(c *gin.Context) {
	if h == nil || h.spuService == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("product service unavailable"))
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("spu id is required"))
		return
	}
	spu, err := h.spuService.Get(c.Request.Context(), spuID)
	if err != nil {
		respondMiniAppError(c, http.StatusNotFound, err)
		return
	}
	if spu == nil || !strings.EqualFold(strings.TrimSpace(spu.Status), "published") {
		respondMiniAppError(c, http.StatusNotFound, errors.New("product not found"))
		return
	}

	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	coverURL, skuCount, minPrice, maxPrice, currency := h.enrichSingleProduct(c.Request.Context(), tenantUUID, spu.ID, spu.Type)
	subtitle, description := h.loadSPULocaleSnippet(c.Request.Context(), tenantUUID, spu.ID, spu.DefaultLocale)

	resp := miniAppProductDetail{
		ID:           spu.ID,
		Code:         spu.Code,
		Name:         spu.Name,
		Type:         spu.Type,
		Status:       spu.Status,
		CategoryID:   spu.CategoryID,
		CategoryPath: spu.CategoryPath,
		Tags:         append([]string(nil), spu.Tags...),
		CoverURL:     coverURL,
		Subtitle:     subtitle,
		Description:  description,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		Currency:     currency,
		PriceLabel:   buildPriceLabel(spu.Type, minPrice, maxPrice, currency),
		SKUCount:     skuCount,
	}
	contracts.ResponseSuccess(c, resp)
}

// GetProductDetail returns SPU + specs + SKU mappings for spec selection.
func (h *Handler) GetProductDetail(c *gin.Context) {
	if h == nil || h.spuService == nil || h.db == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("product service unavailable"))
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("spu id is required"))
		return
	}
	spu, err := h.spuService.Get(c.Request.Context(), spuID)
	if err != nil {
		respondMiniAppError(c, http.StatusNotFound, err)
		return
	}
	if spu == nil || !strings.EqualFold(strings.TrimSpace(spu.Status), "published") {
		respondMiniAppError(c, http.StatusNotFound, errors.New("product not found"))
		return
	}

	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	coverURL, skuCount, minPrice, maxPrice, currency := h.enrichSingleProduct(c.Request.Context(), tenantUUID, spu.ID, spu.Type)
	subtitle, description := h.loadSPULocaleSnippet(c.Request.Context(), tenantUUID, spu.ID, spu.DefaultLocale)

	spuResp := miniAppProductDetail{
		ID:           spu.ID,
		Code:         spu.Code,
		Name:         spu.Name,
		Type:         spu.Type,
		Status:       spu.Status,
		CategoryID:   spu.CategoryID,
		CategoryPath: spu.CategoryPath,
		Tags:         append([]string(nil), spu.Tags...),
		CoverURL:     coverURL,
		Subtitle:     subtitle,
		Description:  description,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		Currency:     currency,
		PriceLabel:   buildPriceLabel(spu.Type, minPrice, maxPrice, currency),
		SKUCount:     skuCount,
	}

	specBundle, groupByID, optionByID := h.loadSpecBundle(c.Request.Context(), tenantUUID, spu.ID)
	skus := h.loadSkuDetailsForSpec(c.Request.Context(), tenantUUID, spu.ID, groupByID, optionByID)

	contracts.ResponseSuccess(c, miniAppProductDetailBundle{
		SPU:  spuResp,
		Spec: specBundle,
		SKUs: skus,
	})
}

func (h *Handler) loadSpecBundle(ctx context.Context, tenantUUID, spuID string) (miniAppProductSpecBundle, map[string]productspecmodel.ProductSpecGroup, map[string]productspecmodel.ProductSpecOption) {
	out := miniAppProductSpecBundle{Groups: []miniAppSpecGroup{}}
	groupByID := map[string]productspecmodel.ProductSpecGroup{}
	optionByID := map[string]productspecmodel.ProductSpecOption{}
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" || strings.TrimSpace(spuID) == "" {
		return out, groupByID, optionByID
	}
	var groups []productspecmodel.ProductSpecGroup
	if err := h.db.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ? AND deleted_at IS NULL AND status = 'active'", tenantUUID, spuID).
		Order("sort_order ASC, code ASC, created_at ASC").
		Find(&groups).Error; err != nil {
		return out, groupByID, optionByID
	}
	groupIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		groupByID[g.ID] = g
		groupIDs = append(groupIDs, g.ID)
	}
	var options []productspecmodel.ProductSpecOption
	if len(groupIDs) > 0 {
		if err := h.db.WithContext(ctx).
			Where("tenant_uuid = ? AND group_id IN ? AND deleted_at IS NULL AND status = 'active'", tenantUUID, groupIDs).
			Order("sort_order ASC, code ASC, created_at ASC").
			Find(&options).Error; err != nil {
			return out, groupByID, optionByID
		}
	}
	optsByGroup := make(map[string][]productspecmodel.ProductSpecOption, len(groupIDs))
	for _, o := range options {
		optionByID[o.ID] = o
		optsByGroup[o.GroupID] = append(optsByGroup[o.GroupID], o)
	}
	for _, g := range groups {
		dto := miniAppSpecGroup{
			ID:        g.ID,
			Code:      g.Code,
			Name:      g.Name,
			Required:  g.Required,
			SortOrder: g.SortOrder,
			Options:   []miniAppSpecOption{},
		}
		for _, o := range optsByGroup[g.ID] {
			dto.Options = append(dto.Options, miniAppSpecOption{
				ID:        o.ID,
				Code:      o.Code,
				Name:      o.Name,
				SortOrder: o.SortOrder,
				Meta:      decodeJSONMap(o.Meta),
			})
		}
		out.Groups = append(out.Groups, dto)
	}
	return out, groupByID, optionByID
}

func decodeJSONMap(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func (h *Handler) loadSkuDetailsForSpec(ctx context.Context, tenantUUID, spuID string, groupByID map[string]productspecmodel.ProductSpecGroup, optionByID map[string]productspecmodel.ProductSpecOption) []miniAppSkuDetail {
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" || strings.TrimSpace(spuID) == "" {
		return []miniAppSkuDetail{}
	}
	tx, done := h.beginTenantTx(ctx, tenantUUID)
	if tx == nil {
		return []miniAppSkuDetail{}
	}
	defer done()
	type skuRow struct {
		ID            string         `gorm:"column:id"`
		SKUCode       string         `gorm:"column:sku_code"`
		Status        string         `gorm:"column:status"`
		SpecValues    datatypes.JSON `gorm:"column:spec_values"`
		DefaultValues datatypes.JSON `gorm:"column:default_values"`
		SpecSignature string         `gorm:"column:spec_signature"`
	}
	var rows []skuRow
	if err := tx.
		Table(productskumodel.ProductSKU{}.TableName()).
		Select("id, sku_code, status, spec_values, default_values, spec_signature").
		Where("tenant_uuid = ? AND spu_id = ? AND deleted_at IS NULL AND status IN ?", tenantUUID, spuID, miniAppVisibleSKUStatuses).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return []miniAppSkuDetail{}
	}
	skuIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		skuIDs = append(skuIDs, r.ID)
	}
	mediaBySKU := h.loadSkuCoverURLsTx(tx, tenantUUID, skuIDs)
	priceBySKU := h.loadSkuPricesTx(tx, tenantUUID, skuIDs)
	stockBySKU := h.loadSkuStockQtyTx(tx, tenantUUID, skuIDs)

	out := make([]miniAppSkuDetail, 0, len(rows))
	for _, r := range rows {
		p := priceBySKU[r.ID]
		spec := map[string]string{}
		var specs []skuservice.SkuSpec
		if len(r.SpecValues) > 0 {
			_ = json.Unmarshal(r.SpecValues, &specs)
		}
		for _, s := range specs {
			gid := strings.TrimSpace(s.SpecID)
			oid := strings.TrimSpace(s.ValueID)
			g, okG := groupByID[gid]
			o, okO := optionByID[oid]
			if !okG || !okO {
				continue
			}
			if strings.TrimSpace(g.Code) == "" || strings.TrimSpace(o.Code) == "" {
				continue
			}
			spec[g.Code] = o.Code
		}
		out = append(out, miniAppSkuDetail{
			ID:            r.ID,
			Code:          r.SKUCode,
			Price:         p.Price,
			Currency:      strings.TrimSpace(p.Currency),
			ImageURL:      strings.TrimSpace(mediaBySKU[r.ID]),
			StockQty:      stockBySKU[r.ID],
			Spec:          spec,
			SpecSignature: strings.TrimSpace(r.SpecSignature),
		})
	}
	return out
}

// ListSkus returns SPU scoped SKU summaries.
func (h *Handler) ListSkus(c *gin.Context) {
	if h == nil || h.skuService == nil || !h.skuService.Ready() {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("sku service unavailable"))
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("spu id is required"))
		return
	}
	var query miniAppSkuListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondMiniAppError(c, http.StatusBadRequest, err)
		return
	}
	status := strings.TrimSpace(query.Status)
	if status == "" {
		status = "online"
	}
	if strings.EqualFold(status, "published") {
		// Backward compatible alias; canonical SKU status is "online".
		status = "online"
	}
	if !strings.EqualFold(status, "online") {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("only online skus are accessible via mini-app"))
		return
	}
	result, err := h.skuService.ListSkus(c.Request.Context(), skuservice.SkuListQuery{
		SPUID:    spuID,
		Status:   status,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		respondMiniAppError(c, http.StatusInternalServerError, err)
		return
	}
	if result == nil {
		result = &skuservice.SkuListResult{
			Items:    []skuservice.SkuListItem{},
			Page:     query.Page,
			PageSize: query.PageSize,
		}
	}

	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	enriched := h.enrichSkuSummaries(c.Request.Context(), tenantUUID, result.Items)

	resp := skuListResponse{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Items:    make([]miniAppSkuSummary, 0, len(enriched)),
	}
	resp.Items = append(resp.Items, enriched...)
	contracts.ResponseSuccess(c, resp)
}

// ListSubscriptionPlans returns active subscription plans for a subscription SPU.
func (h *Handler) ListSubscriptionPlans(c *gin.Context) {
	if h == nil || h.planSvc == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("subscription plan service unavailable"))
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("spu id is required"))
		return
	}
	plans, err := h.planSvc.List(c.Request.Context(), spuID)
	if err != nil {
		respondMiniAppError(c, http.StatusInternalServerError, err)
		return
	}
	items := make([]miniAppSubscriptionPlan, 0, len(plans))
	for _, p := range plans {
		if !strings.EqualFold(strings.TrimSpace(p.Status), "active") {
			continue
		}
		items = append(items, miniAppSubscriptionPlan{
			ID:           p.ID,
			PlanCode:     p.PlanCode,
			Name:         p.Name,
			BillingCycle: p.BillingCycle,
			BillingValue: p.BillingValue,
			Price:        p.Price,
			Currency:     p.Currency,
			TrialDays:    p.TrialDays,
			AutoRenew:    p.AutoRenew,
			CancelPolicy: p.CancelPolicy,
			Status:       p.Status,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Currency == items[j].Currency {
			return items[i].Price < items[j].Price
		}
		return items[i].Currency < items[j].Currency
	})
	contracts.ResponseSuccess(c, subscriptionPlanListResponse{Items: items})
}

type skuRow struct {
	ID            string         `gorm:"column:id" json:"id"`
	SPUID         string         `gorm:"column:spu_id" json:"spu_id"`
	DefaultValues datatypes.JSON `gorm:"column:default_values" json:"default_values"`
}

type skuMediaRow struct {
	SPUID     string
	SKUID     string
	URL       string
	IsPrimary bool
	SortOrder int
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (h *Handler) enrichProductSummaries(ctx context.Context, tenantUUID string, items []spuservice.SPUSummary) []miniAppProductSummary {
	if len(items) == 0 {
		return []miniAppProductSummary{}
	}
	out := make([]miniAppProductSummary, 0, len(items))
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" {
		for _, item := range items {
			updatedAt := item.UpdatedAt
			out = append(out, miniAppProductSummary{
				ID:        item.ID,
				Code:      item.Code,
				Name:      item.Name,
				Type:      item.Type,
				Status:    item.Status,
				UpdatedAt: &updatedAt,
			})
		}
		return out
	}

	spuIDs := make([]string, 0, len(items))
	typeByID := make(map[string]string, len(items))
	for _, item := range items {
		spuIDs = append(spuIDs, item.ID)
		typeByID[item.ID] = item.Type
	}

	coverBySPU := h.loadCoverURLs(ctx, tenantUUID, spuIDs)
	skuAgg := h.loadSkuAgg(ctx, tenantUUID, spuIDs)
	planAgg := h.loadPlanAgg(ctx, tenantUUID, spuIDs)

	for _, item := range items {
		updatedAt := item.UpdatedAt
		cover := coverBySPU[item.ID]
		agg := skuAgg[item.ID]
		minPrice := agg.MinPrice
		maxPrice := agg.MaxPrice
		currency := agg.Currency
		skuCount := agg.SKUCount
		if strings.EqualFold(strings.TrimSpace(typeByID[item.ID]), "subscription") {
			if p, ok := planAgg[item.ID]; ok {
				minPrice = p.MinPrice
				maxPrice = p.MaxPrice
				currency = p.Currency
				skuCount = p.PlanCount
			}
		}
		out = append(out, miniAppProductSummary{
			ID:         item.ID,
			Code:       item.Code,
			Name:       item.Name,
			Type:       item.Type,
			Status:     item.Status,
			UpdatedAt:  &updatedAt,
			CoverURL:   cover,
			MinPrice:   minPrice,
			MaxPrice:   maxPrice,
			Currency:   currency,
			PriceLabel: buildPriceLabel(item.Type, minPrice, maxPrice, currency),
			SKUCount:   skuCount,
		})
	}
	return out
}

type skuAggInfo struct {
	MinPrice *float64
	MaxPrice *float64
	Currency string
	SKUCount int
}

type planAggInfo struct {
	MinPrice  *float64
	MaxPrice  *float64
	Currency  string
	PlanCount int
}

func (h *Handler) loadSkuAgg(ctx context.Context, tenantUUID string, spuIDs []string) map[string]skuAggInfo {
	out := make(map[string]skuAggInfo, len(spuIDs))
	if h == nil || h.db == nil || len(spuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}
	tx, done := h.beginTenantTx(ctx, tenantUUID)
	if tx == nil {
		return out
	}
	defer done()

	var rows []skuRow
	if err := tx.
		Table(productskumodel.ProductSKU{}.TableName()).
		Select("id, spu_id, default_values").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND status IN ? AND spu_id IN ?", tenantUUID, miniAppVisibleSKUStatuses, spuIDs).
		Find(&rows).Error; err != nil {
		return out
	}

	skuIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		skuIDs = append(skuIDs, r.ID)
	}
	pricebookPrices, pbCurrency := h.loadBasePricebookSkuPricesTx(tx, tenantUUID, skuIDs)

	for _, r := range rows {
		agg := out[r.SPUID]
		agg.SKUCount++
		if p, ok := pricebookPrices[r.ID]; ok && p != nil {
			agg.Currency = pickCurrency(agg.Currency, pbCurrency)
			agg.MinPrice, agg.MaxPrice = mergeMinMax(agg.MinPrice, agg.MaxPrice, *p)
		} else {
			price, currency, ok := extractPriceFromDefaultValues(r.DefaultValues)
			if ok {
				agg.Currency = pickCurrency(agg.Currency, currency)
				agg.MinPrice, agg.MaxPrice = mergeMinMax(agg.MinPrice, agg.MaxPrice, price)
			}
		}
		out[r.SPUID] = agg
	}
	return out
}

func (h *Handler) loadPlanAgg(ctx context.Context, tenantUUID string, spuIDs []string) map[string]planAggInfo {
	out := make(map[string]planAggInfo, len(spuIDs))
	if h == nil || h.db == nil || len(spuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}
	type planRow struct {
		SPUID    string  `gorm:"column:spu_id"`
		Price    float64 `gorm:"column:price"`
		Currency string  `gorm:"column:currency"`
		Status   string  `gorm:"column:status"`
	}
	var rows []planRow
	if err := h.db.WithContext(ctx).
		Table(productmodel.SubscriptionPlan{}.TableName()).
		Select("spu_id, price, currency, status").
		Where("tenant_uuid = ? AND spu_id IN ?", tenantUUID, spuIDs).
		Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		if !strings.EqualFold(strings.TrimSpace(r.Status), "active") {
			continue
		}
		agg := out[r.SPUID]
		agg.PlanCount++
		agg.Currency = pickCurrency(agg.Currency, strings.TrimSpace(r.Currency))
		agg.MinPrice, agg.MaxPrice = mergeMinMax(agg.MinPrice, agg.MaxPrice, r.Price)
		out[r.SPUID] = agg
	}
	return out
}

func (h *Handler) loadCoverURLs(ctx context.Context, tenantUUID string, spuIDs []string) map[string]string {
	out := make(map[string]string, len(spuIDs))
	if h == nil || h.db == nil || len(spuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}
	var rows []skuMediaRow
	tx, done := h.beginTenantTx(ctx, tenantUUID)
	if tx == nil {
		return out
	}
	defer done()

	mediaTable := fmt.Sprintf("%s AS m", productskumodel.ProductSKUMedia{}.TableName())
	skuTable := fmt.Sprintf("%s AS s", productskumodel.ProductSKU{}.TableName())
	if err := tx.
		Table(mediaTable).
		Select("s.spu_id AS spu_id, m.sku_id AS sku_id, m.url, m.is_primary, m.sort_order, m.created_at").
		Joins("JOIN "+skuTable+" ON s.id = m.sku_id").
		Where("m.tenant_uuid = ? AND s.tenant_uuid = ? AND m.deleted_at IS NULL AND s.deleted_at IS NULL AND s.status IN ? AND s.spu_id IN ?",
			tenantUUID, tenantUUID, miniAppVisibleSKUStatuses, spuIDs).
		Order("s.spu_id ASC, m.is_primary DESC, m.sort_order ASC, m.created_at ASC").
		Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		if _, exists := out[r.SPUID]; exists {
			continue
		}
		if strings.TrimSpace(r.URL) == "" {
			continue
		}
		out[r.SPUID] = strings.TrimSpace(r.URL)
	}
	return out
}

func (h *Handler) enrichSingleProduct(ctx context.Context, tenantUUID, spuID, spuType string) (coverURL string, skuCount int, minPrice, maxPrice *float64, currency string) {
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" || strings.TrimSpace(spuID) == "" {
		return "", 0, nil, nil, ""
	}
	cover := h.loadCoverURLs(ctx, tenantUUID, []string{spuID})
	coverURL = cover[spuID]

	if strings.EqualFold(strings.TrimSpace(spuType), "subscription") {
		planAgg := h.loadPlanAgg(ctx, tenantUUID, []string{spuID})
		if p, ok := planAgg[spuID]; ok {
			return coverURL, p.PlanCount, p.MinPrice, p.MaxPrice, p.Currency
		}
		return coverURL, 0, nil, nil, ""
	}
	skuAgg := h.loadSkuAgg(ctx, tenantUUID, []string{spuID})
	if s, ok := skuAgg[spuID]; ok {
		return coverURL, s.SKUCount, s.MinPrice, s.MaxPrice, s.Currency
	}
	return coverURL, 0, nil, nil, ""
}

func (h *Handler) loadSPULocaleSnippet(ctx context.Context, tenantUUID, spuID, locale string) (subtitle, description string) {
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" || strings.TrimSpace(spuID) == "" {
		return "", ""
	}
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = "zh-CN"
	}
	type row struct {
		Subtitle    string `gorm:"column:subtitle"`
		Description string `gorm:"column:description"`
		Status      string `gorm:"column:status"`
	}
	var r row
	err := h.db.WithContext(ctx).
		Table(productmodel.SPULocale{}.TableName()).
		Select("subtitle, description, status").
		Where("tenant_uuid = ? AND spu_id = ? AND locale = ?", tenantUUID, spuID, locale).
		First(&r).Error
	if err != nil {
		return "", ""
	}
	if strings.EqualFold(strings.TrimSpace(r.Status), "inactive") {
		return "", ""
	}
	return strings.TrimSpace(r.Subtitle), strings.TrimSpace(r.Description)
}

func (h *Handler) enrichSkuSummaries(ctx context.Context, tenantUUID string, items []skuservice.SkuListItem) []miniAppSkuSummary {
	if len(items) == 0 {
		return []miniAppSkuSummary{}
	}
	out := make([]miniAppSkuSummary, 0, len(items))
	if h == nil || h.db == nil || strings.TrimSpace(tenantUUID) == "" {
		for _, item := range items {
			out = append(out, toMiniAppSkuSummary(item, "", nil, "", nil))
		}
		return out
	}
	ids := make([]string, 0, len(items))
	spuID := ""
	for _, it := range items {
		ids = append(ids, it.ID)
		spuID = it.SPUID
	}

	tx, done := h.beginTenantTx(ctx, tenantUUID)
	if tx == nil {
		for _, item := range items {
			out = append(out, toMiniAppSkuSummary(item, "", nil, "", nil))
		}
		return out
	}
	defer done()

	mediaBySKU := h.loadSkuCoverURLsTx(tx, tenantUUID, ids)
	priceBySKU := h.loadSkuPricesTx(tx, tenantUUID, ids)
	stockBySKU := h.loadSkuStockQtyTx(tx, tenantUUID, ids)

	for _, item := range items {
		p := priceBySKU[item.ID]
		out = append(out, toMiniAppSkuSummary(item, mediaBySKU[item.ID], p.Price, p.Currency, stockBySKU[item.ID]))
	}
	_ = spuID
	return out
}

type skuPriceInfo struct {
	Price    *float64
	Currency string
}

func (h *Handler) loadSkuPricesTx(tx *gorm.DB, tenantUUID string, skuIDs []string) map[string]skuPriceInfo {
	out := make(map[string]skuPriceInfo, len(skuIDs))
	if h == nil || tx == nil || len(skuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}

	// Prefer base pricebook (if items exist) so mini-app can reflect pricing rules.
	pricebookPrices, pbCurrency := h.loadBasePricebookSkuPricesTx(tx, tenantUUID, skuIDs)

	var rows []skuRow
	if err := tx.
		Table(productskumodel.ProductSKU{}.TableName()).
		Select("id, spu_id, default_values").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND id IN ?", tenantUUID, skuIDs).
		Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		if p, ok := pricebookPrices[r.ID]; ok {
			out[r.ID] = skuPriceInfo{Price: p, Currency: pbCurrency}
			continue
		}
		price, currency, ok := extractPriceFromDefaultValues(r.DefaultValues)
		if !ok {
			continue
		}
		p := price
		out[r.ID] = skuPriceInfo{Price: &p, Currency: currency}
	}
	return out
}

func (h *Handler) loadBasePricebookSkuPricesTx(tx *gorm.DB, tenantUUID string, skuIDs []string) (map[string]*float64, string) {
	out := make(map[string]*float64, len(skuIDs))
	if h == nil || tx == nil || len(skuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out, ""
	}

	type pbRow struct {
		ID             string  `gorm:"column:id"`
		Currency       string  `gorm:"column:currency"`
		CurrentVersion *string `gorm:"column:current_version_id"`
	}
	var pb pbRow
	if err := tx.Table(pricingmodel.Pricebook{}.TableName()).
		Select("id, currency, current_version_id").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND code = ? AND status = ?", tenantUUID, "base", "active").
		First(&pb).Error; err != nil {
		return out, ""
	}
	versionID := ""
	if pb.CurrentVersion != nil {
		versionID = strings.TrimSpace(*pb.CurrentVersion)
	}
	if versionID == "" {
		return out, strings.TrimSpace(pb.Currency)
	}

	type itemRow struct {
		SKUID           string `gorm:"column:sku_id"`
		BaseAmountMinor *int64 `gorm:"column:base_amount_minor"`
		SaleAmountMinor *int64 `gorm:"column:sale_amount_minor"`
	}
	var items []itemRow
	if err := tx.Table(pricingmodel.PricebookItem{}.TableName()).
		Select("sku_id, base_amount_minor, sale_amount_minor").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND version_id = ? AND sku_id IN ?", tenantUUID, versionID, skuIDs).
		Find(&items).Error; err != nil {
		return out, strings.TrimSpace(pb.Currency)
	}

	for _, it := range items {
		minor := (*int64)(nil)
		if it.SaleAmountMinor != nil {
			minor = it.SaleAmountMinor
		} else if it.BaseAmountMinor != nil {
			minor = it.BaseAmountMinor
		}
		if minor == nil {
			continue
		}
		v := float64(*minor) / 100.0
		out[strings.TrimSpace(it.SKUID)] = &v
	}
	return out, strings.TrimSpace(pb.Currency)
}

func (h *Handler) loadSkuCoverURLsTx(tx *gorm.DB, tenantUUID string, skuIDs []string) map[string]string {
	out := make(map[string]string, len(skuIDs))
	if h == nil || tx == nil || len(skuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}
	type row struct {
		SKUID     string    `gorm:"column:sku_id"`
		URL       string    `gorm:"column:url"`
		IsPrimary bool      `gorm:"column:is_primary"`
		SortOrder int       `gorm:"column:sort_order"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	if err := tx.
		Table(productskumodel.ProductSKUMedia{}.TableName()).
		Select("sku_id, url, is_primary, sort_order, created_at").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND sku_id IN ?", tenantUUID, skuIDs).
		Order("sku_id ASC, is_primary DESC, sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		if _, exists := out[r.SKUID]; exists {
			continue
		}
		if strings.TrimSpace(r.URL) == "" {
			continue
		}
		out[r.SKUID] = strings.TrimSpace(r.URL)
	}
	return out
}

func (h *Handler) loadSkuStockQtyTx(tx *gorm.DB, tenantUUID string, skuIDs []string) map[string]*int {
	out := make(map[string]*int, len(skuIDs))
	if h == nil || tx == nil || len(skuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out
	}
	type row struct {
		SKUID string `gorm:"column:sku_id"`
		Qty   int64  `gorm:"column:qty"`
	}
	var rows []row
	if err := tx.Table(productskumodel.ProductSKUInventory{}.TableName()).
		Select("sku_id, SUM(available_qty) AS qty").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND sku_id IN ?", tenantUUID, skuIDs).
		Group("sku_id").
		Scan(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		id := strings.TrimSpace(r.SKUID)
		if id == "" {
			continue
		}
		out[id] = clampInt64ToIntPtr(r.Qty)
	}
	return out
}

func toMiniAppSkuSummary(item skuservice.SkuListItem, imageURL string, price *float64, currency string, stockQty *int) miniAppSkuSummary {
	var createdAt, updatedAt *time.Time
	if item.CreatedAt != nil {
		c := *item.CreatedAt
		createdAt = &c
	}
	if item.UpdatedAt != nil {
		u := *item.UpdatedAt
		updatedAt = &u
	}
	return miniAppSkuSummary{
		ID:        item.ID,
		SPUID:     item.SPUID,
		Code:      item.SKUCode,
		Status:    item.Status,
		Barcode:   item.Barcode,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ImageURL:  strings.TrimSpace(imageURL),
		Price:     price,
		Currency:  strings.TrimSpace(currency),
		StockQty:  stockQty,
	}
}

func extractPriceFromDefaultValues(raw datatypes.JSON) (price float64, currency string, ok bool) {
	if len(raw) == 0 {
		return 0, "", false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, "", false
	}
	currency = strings.TrimSpace(asString(m["currency"]))
	if currency == "" {
		currency = "CNY"
	}
	if v, ok2 := asFloat(m["sale_price"]); ok2 {
		return v, currency, true
	}
	if v, ok2 := asFloat(m["price"]); ok2 {
		return v, currency, true
	}
	if v, ok2 := asFloat(m["salePrice"]); ok2 {
		return v, currency, true
	}
	if v, ok2 := asFloat(m["list_price"]); ok2 {
		return v, currency, true
	}
	return 0, "", false
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), "\"")
	}
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		n := json.Number(s)
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func mergeMinMax(min, max *float64, v float64) (*float64, *float64) {
	if min == nil || v < *min {
		n := v
		min = &n
	}
	if max == nil || v > *max {
		n := v
		max = &n
	}
	return min, max
}

func pickCurrency(existing, incoming string) string {
	existing = strings.TrimSpace(existing)
	incoming = strings.TrimSpace(incoming)
	if existing != "" {
		return existing
	}
	return incoming
}

func buildPriceLabel(spuType string, minPrice, maxPrice *float64, currency string) string {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		currency = "CNY"
	}
	symbol := currencySymbol(currency)
	if minPrice == nil && maxPrice == nil {
		if strings.EqualFold(strings.TrimSpace(spuType), "subscription") {
			return symbol + "--/期"
		}
		return symbol + "--"
	}
	if maxPrice != nil && minPrice != nil && *maxPrice != *minPrice {
		if strings.EqualFold(strings.TrimSpace(spuType), "subscription") {
			return symbol + formatPrice(*minPrice) + "起/期"
		}
		return symbol + formatPrice(*minPrice) + " 起"
	}
	v := minPrice
	if v == nil {
		v = maxPrice
	}
	if v == nil {
		return symbol + "--"
	}
	if strings.EqualFold(strings.TrimSpace(spuType), "subscription") {
		return symbol + formatPrice(*v) + "/期"
	}
	return symbol + formatPrice(*v)
}

func currencySymbol(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "CNY", "RMB":
		return "¥"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "HKD":
		return "HK$"
	default:
		return currency + " "
	}
}

func formatPrice(v float64) string {
	s := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
	if s == "" {
		return "0"
	}
	return s
}

func respondMiniAppError(c *gin.Context, status int, err error) {
	if status == 0 {
		status = statusFromError(err)
	}
	if err == nil {
		err = errors.New("unknown error")
	}
	message := err.Error()
	switch status {
	case http.StatusBadRequest:
		contracts.ResponseBadRequest(c, message)
	case http.StatusUnauthorized:
		contracts.ResponseUnauthorized(c, message)
	case http.StatusNotFound:
		contracts.ResponseNotFound(c, message)
	case http.StatusServiceUnavailable:
		contracts.ResponseServiceUnavailable(c, message, nil)
	default:
		contracts.ResponseError(c, status, contracts.ErrCodeInternalError, message)
	}
}

func statusFromError(err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "tenant"):
		return http.StatusUnauthorized
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"),
		strings.Contains(msg, "invalid"),
		strings.Contains(msg, "missing"),
		strings.Contains(msg, "spu id"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
