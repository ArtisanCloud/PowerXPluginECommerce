package product_category

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
)

type templateSimulateRequest struct {
	CategoryID string         `json:"categoryId,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	TemplateID string         `json:"templateId,omitempty"`
}

func (h *Handler) TemplatePreview(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	// 预览：返回草稿模板详情（含 fields）
	detail, err := h.service.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func (h *Handler) TemplateSimulate(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req templateSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	// 模拟校验：按类目取生效模板（若传 templateId 则以该模板为准）并校验 attributes
	var vErrs categoryservice.ValidationErrors
	var err error
	switch {
	case req.TemplateID != "":
		detail, getErr := h.service.GetTemplate(c.Request.Context(), req.TemplateID)
		if getErr != nil {
			respondTemplateError(c, getErr)
			return
		}
		vErrs = h.service.ValidateAttributesPreview(detail.Fields, req.Attributes)
	case req.CategoryID != "":
		vErrs, err = h.service.ValidateAttributesForCategory(c.Request.Context(), req.CategoryID, req.Attributes)
		if err != nil {
			respondTemplateError(c, err)
			return
		}
	default:
		contracts.ResponseBadRequest(c, "categoryId or templateId is required")
		return
	}

	if len(vErrs) > 0 {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"ok": true}, "validation passed")
}
