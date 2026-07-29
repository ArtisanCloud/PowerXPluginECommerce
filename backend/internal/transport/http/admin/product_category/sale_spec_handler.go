package product_category

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) SaleSpecList(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	categoryID := strings.TrimSpace(c.Param("id"))
	if categoryID == "" {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "category id is required")
		return
	}
	groups, err := h.service.ListSaleSpecs(c.Request.Context(), categoryID)
	if err != nil {
		respondSaleSpecError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"groups": groups})
}

func (h *Handler) SaleSpecReplace(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	categoryID := strings.TrimSpace(c.Param("id"))
	if categoryID == "" {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "category id is required")
		return
	}
	var req categoryservice.ReplaceSaleSpecsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	groups, err := h.service.ReplaceSaleSpecs(c.Request.Context(), categoryID, req)
	if err != nil {
		respondSaleSpecError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"groups": groups}, "category sale specs updated")
}

func respondSaleSpecError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, "unknown error")
		return
	}
	if vErrs, ok := err.(categoryservice.ValidationErrors); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "validation failed", vErrs)
		return
	}
	if err == gorm.ErrRecordNotFound {
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "category not found")
		return
	}
	contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
}
