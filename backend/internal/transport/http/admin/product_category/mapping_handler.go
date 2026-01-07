package product_category

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) MappingList(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	items, err := h.service.ListMappings(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondMappingError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": items})
}

func (h *Handler) MappingUpsert(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.UpsertMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	result, err := h.service.UpsertMapping(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondMappingError(c, err)
		return
	}
	if strings.EqualFold(strings.TrimSpace(req.Operation), "delete") {
		contracts.ResponseSuccessWithMessage(c, gin.H{"ok": true}, "mapping deleted")
		return
	}
	contracts.ResponseSuccessWithMessage(c, result, "mapping saved")
}

func respondMappingError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, "unknown error")
		return
	}
	if vErrs, ok := err.(categoryservice.ValidationErrors); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
		return
	}
	if cErr, ok := err.(*categoryservice.ConflictError); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusConflict, contracts.ErrCodeConflict, cErr.Message, cErr)
		return
	}
	if err == gorm.ErrRecordNotFound {
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "record not found")
		return
	}
	contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
}
