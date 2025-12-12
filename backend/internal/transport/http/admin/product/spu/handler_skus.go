package spu

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// SKUHandler exposes HTTP endpoints for SKU linking APIs.
type SKUHandler struct {
	service *spuservice.SKULinkService
}

// NewSKUHandler constructs a SKU handler.
func NewSKUHandler(service *spuservice.SKULinkService) *SKUHandler {
	return &SKUHandler{service: service}
}

// List returns SKU associations for an SPU.
func (h *SKUHandler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "sku service unavailable")
		return
	}
	skus, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": skus})
}

// Replace overwrites SKU associations for an SPU.
func (h *SKUHandler) Replace(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "sku service unavailable")
		return
	}
	var req spuservice.ReplaceSKUsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	result, err := h.service.Replace(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"items": result}, "sku links updated")
}
