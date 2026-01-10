package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	productModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productsku "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ItemsHandler struct {
	items *pricingsvc.ItemService
}

func NewItemsHandler(domain *pricingsvc.Service) *ItemsHandler {
	if domain == nil || !domain.Ready() {
		return &ItemsHandler{items: nil}
	}
	return &ItemsHandler{items: pricingsvc.NewItemService(domain.Deps())}
}

func (h *ItemsHandler) List(c *gin.Context) {
	if h == nil || h.items == nil || !h.items.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	versionID := strings.TrimSpace(c.Param("versionId"))
	if pricebookID == "" || versionID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId/versionId are required"))
		return
	}

	page := intFromQuery(c, "page", 1)
	pageSize := intFromQuery(c, "page_size", 200)

	res, err := h.items.ListItems(c.Request.Context(), pricingsvc.ListItemsInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		SKUID:       c.Query("sku_id"),
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	tenantUUID, terr := authx.RequireTenantUUID(c.Request.Context())
	if terr != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeTenantMissing, terr)
		return
	}
	locale := strings.TrimSpace(c.Query("locale"))
	if locale == "" {
		locale = "zh-CN"
	}

	skuIDs := make([]string, 0, len(res.Items))
	for _, it := range res.Items {
		if it == nil || strings.TrimSpace(it.SKUID) == "" {
			continue
		}
		skuIDs = append(skuIDs, strings.TrimSpace(it.SKUID))
	}
	skuByID, spuNameByID := loadSkuAndSpuNames(c.Request.Context(), h.items.Deps().DB, tenantUUID, skuIDs, locale)

	items := make([]PricebookItemDTO, 0, len(res.Items))
	for _, it := range res.Items {
		var sku *productsku.ProductSKU
		if it != nil {
			if s, ok := skuByID[strings.TrimSpace(it.SKUID)]; ok {
				ss := s
				sku = &ss
			}
		}
		items = append(items, toItemDTO(it, sku, spuNameByID))
	}
	contracts.ResponseSuccess(c, ItemsListResponse{
		Items: items,
		Meta:  PageMeta{Page: res.Page, PageSize: res.PageSize, Total: res.Total},
	})
}

func (h *ItemsHandler) Upsert(c *gin.Context) {
	if h == nil || h.items == nil || !h.items.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	versionID := strings.TrimSpace(c.Param("versionId"))
	if pricebookID == "" || versionID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId/versionId are required"))
		return
	}

	var req ItemsUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	items := make([]pricingsvc.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, pricingsvc.ItemInput{
			SKUID:           it.SKUID,
			BaseAmountMinor: it.BaseAmountMinor,
			SaleAmountMinor: it.SaleAmountMinor,
			MsrpAmountMinor: it.MsrpAmountMinor,
			CostAmountMinor: it.CostAmountMinor,
			MinAmountMinor:  it.MinAmountMinor,
			MaxAmountMinor:  it.MaxAmountMinor,
			TaxIncluded:     it.TaxIncluded,
			Meta:            it.Meta,
		})
	}

	res, err := h.items.UpsertItems(c.Request.Context(), pricingsvc.UpsertItemsInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		Items:       items,
		Actor:       actorFromContext(c),
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, ItemsUpsertResponse{
		Upserted: res.Upserted,
		Skipped:  res.Skipped,
	})
}

func toItemDTO(it *pricingModel.PricebookItem, sku *productsku.ProductSKU, spuNameByID map[string]string) PricebookItemDTO {
	if it == nil {
		return PricebookItemDTO{}
	}
	var meta map[string]any
	if len(it.Meta) > 0 {
		_ = json.Unmarshal(it.Meta, &meta)
	}
	createdAt := it.CreatedAt
	updatedAt := it.UpdatedAt
	specDisplay := strings.TrimSpace(asString(meta["spec_display"]))
	skuCode := strings.TrimSpace(asString(meta["sku_code"]))
	spuID := ""
	spuName := ""
	if sku != nil {
		if skuCode == "" {
			skuCode = strings.TrimSpace(sku.SKUCode)
		}
		spuID = strings.TrimSpace(sku.SPUID)
		if specDisplay == "" {
			specDisplay = buildSpecDisplayFromPairs(sku.SpecValues)
		}
	}
	if spuID != "" {
		spuName = strings.TrimSpace(spuNameByID[spuID])
	}
	return PricebookItemDTO{
		ID:              it.ID,
		PricebookID:     it.PricebookID,
		VersionID:       it.VersionID,
		SKUID:           it.SKUID,
		SKUCode:         skuCode,
		SPUID:           spuID,
		SPUName:         spuName,
		SpecDisplay:     specDisplay,
		BaseAmountMinor: it.BaseAmount,
		SaleAmountMinor: it.SaleAmount,
		MsrpAmountMinor: it.MsrpAmount,
		CostAmountMinor: it.CostAmount,
		MinAmountMinor:  it.MinAmount,
		MaxAmountMinor:  it.MaxAmount,
		TaxIncluded:     it.TaxIncluded,
		Meta:            meta,
		CreatedAt:       &createdAt,
		UpdatedAt:       &updatedAt,
	}
}

func loadSkuAndSpuNames(ctx context.Context, db *gorm.DB, tenantUUID string, skuIDs []string, locale string) (map[string]productsku.ProductSKU, map[string]string) {
	skuByID := map[string]productsku.ProductSKU{}
	spuNameByID := map[string]string{}
	if db == nil || ctx == nil || strings.TrimSpace(tenantUUID) == "" || len(skuIDs) == 0 {
		return skuByID, spuNameByID
	}
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return skuByID, spuNameByID
	}
	defer func() { _ = tx.Rollback() }()
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		_ = tx.Exec("SELECT set_config('app.tenant_uuid', ?, true)", strings.TrimSpace(tenantUUID)).Error
	}

	var skus []productsku.ProductSKU
	if err := tx.Where("tenant_uuid = ? AND deleted_at IS NULL AND id IN ?", tenantUUID, skuIDs).Find(&skus).Error; err != nil {
		return skuByID, spuNameByID
	}
	spuIDs := make([]string, 0, len(skus))
	for _, s := range skus {
		skuByID[strings.TrimSpace(s.ID)] = s
		if strings.TrimSpace(s.SPUID) != "" {
			spuIDs = append(spuIDs, strings.TrimSpace(s.SPUID))
		}
	}
	if len(spuIDs) == 0 {
		return skuByID, spuNameByID
	}

	if tx.Migrator() != nil && tx.Migrator().HasTable(&productModel.SPULocale{}) {
		type locRow struct {
			SPUID  string `gorm:"column:spu_id"`
			Title  string `gorm:"column:title"`
			Status string `gorm:"column:status"`
		}
		var locs []locRow
		_ = tx.Table(productModel.SPULocale{}.TableName()).
			Select("spu_id, title, status").
			Where("tenant_uuid = ? AND spu_id IN ? AND locale = ?", tenantUUID, spuIDs, strings.TrimSpace(locale)).
			Find(&locs).Error
		for _, r := range locs {
			if strings.EqualFold(strings.TrimSpace(r.Status), "inactive") {
				continue
			}
			if strings.TrimSpace(r.Title) == "" {
				continue
			}
			spuNameByID[strings.TrimSpace(r.SPUID)] = strings.TrimSpace(r.Title)
		}
	}

	var spus []productModel.SPU
	if err := tx.Where("tenant_uuid = ? AND deleted_at IS NULL AND id IN ?", tenantUUID, spuIDs).Find(&spus).Error; err == nil {
		for _, s := range spus {
			id := strings.TrimSpace(s.ID)
			if id == "" {
				continue
			}
			if _, ok := spuNameByID[id]; ok {
				continue
			}
			if strings.TrimSpace(s.Name) == "" {
				continue
			}
			spuNameByID[id] = strings.TrimSpace(s.Name)
		}
	}

	return skuByID, spuNameByID
}

type skuSpecJSON struct {
	SpecName  string `json:"spec_name,omitempty"`
	ValueName string `json:"value_name,omitempty"`
}

func buildSpecDisplayFromPairs(raw datatypes.JSON) string {
	if len(raw) == 0 {
		return ""
	}
	var pairs []skuSpecJSON
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return ""
	}
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		k := strings.TrimSpace(p.SpecName)
		v := strings.TrimSpace(p.ValueName)
		if k == "" || v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, " | ")
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), "\"")
	}
}
