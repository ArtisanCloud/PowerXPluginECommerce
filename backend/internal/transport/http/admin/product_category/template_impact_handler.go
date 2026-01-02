package product_category

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
)

func (h *Handler) TemplateImpact(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	summary, err := h.service.GetTemplateImpact(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, summary)
}

func (h *Handler) TemplateTriggerRecheck(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.TemplateRecheckRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
	}
	result, err := h.service.TriggerTemplateRecheck(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, result, "recheck scheduled")
}
