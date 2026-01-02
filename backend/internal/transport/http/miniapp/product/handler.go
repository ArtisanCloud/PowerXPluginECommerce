package product

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	skuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// Handler exposes read-only product/SKU endpoints to mini-app clients.
type Handler struct {
	spuService *spuservice.Service
	skuService *skuservice.Service
}

// NewHandler wires SPU/SKU services for handlers.
func NewHandler(spuSvc *spuservice.Service, skuSvc *skuservice.Service) *Handler {
	return &Handler{
		spuService: spuSvc,
		skuService: skuSvc,
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
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Q)
	}
	result, err := h.spuService.List(c.Request.Context(), spuservice.ListFilters{
		Keyword:            keyword,
		Status:             strings.TrimSpace(query.Status),
		Type:               strings.TrimSpace(query.Type),
		CategoryID:         strings.TrimSpace(query.CategoryID),
		CategoryPathPrefix: strings.TrimSpace(query.CategoryPathPrefix),
		Page:               query.Page,
		PageSize:           query.PageSize,
	})
	if err != nil {
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
	resp := productListResponse{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Items:    make([]miniAppProductSummary, 0, len(result.Items)),
	}
	for _, item := range result.Items {
		updatedAt := item.UpdatedAt
		resp.Items = append(resp.Items, miniAppProductSummary{
			ID:        item.ID,
			Code:      item.Code,
			Name:      item.Name,
			Type:      item.Type,
			Status:    item.Status,
			UpdatedAt: &updatedAt,
		})
	}
	contracts.ResponseSuccess(c, resp)
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
	result, err := h.skuService.ListSkus(c.Request.Context(), skuservice.SkuListQuery{
		SPUID:    spuID,
		Status:   strings.TrimSpace(query.Status),
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
	resp := skuListResponse{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Items:    make([]miniAppSkuSummary, 0, len(result.Items)),
	}
	for _, item := range result.Items {
		var createdAt, updatedAt *time.Time
		if item.CreatedAt != nil {
			c := *item.CreatedAt
			createdAt = &c
		}
		if item.UpdatedAt != nil {
			u := *item.UpdatedAt
			updatedAt = &u
		}
		resp.Items = append(resp.Items, miniAppSkuSummary{
			ID:        item.ID,
			SPUID:     item.SPUID,
			Code:      item.SKUCode,
			Status:    item.Status,
			Barcode:   item.Barcode,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	contracts.ResponseSuccess(c, resp)
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
