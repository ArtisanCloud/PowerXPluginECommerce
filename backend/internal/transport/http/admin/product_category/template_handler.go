package product_category

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) TemplateList(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	result, err := h.service.ListTemplates(c.Request.Context(), categoryservice.TemplateListFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     toInt(c.Query("page"), 1),
		PageSize: toInt(c.Query("pageSize"), 50),
	})
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"items": result.Items,
		"meta":  gin.H{"total": result.Total, "page": result.Page, "pageSize": result.PageSize},
	})
}

func (h *Handler) TemplateGet(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	detail, err := h.service.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func (h *Handler) TemplateCreate(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	created, err := h.service.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, created, "template created")
}

func (h *Handler) TemplateUpdate(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	updated, err := h.service.UpdateTemplate(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, updated, "template updated")
}

func (h *Handler) TemplatePublish(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.TemplatePublishRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
	}
	summary, err := h.service.PublishTemplate(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "template published")
}

func (h *Handler) TemplateRollback(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.TemplateRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	summary, err := h.service.RollbackTemplate(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "template rollback published")
}

func (h *Handler) TemplateVersions(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	items, err := h.service.ListTemplateVersions(c.Request.Context(), c.Param("id"), toInt(c.Query("limit"), 50))
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": items})
}

func (h *Handler) TemplateEffective(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	categoryID := strings.TrimSpace(c.Query("categoryId"))
	if categoryID == "" {
		contracts.ResponseBadRequest(c, "categoryId is required")
		return
	}
	effective, err := h.service.GetEffectiveTemplateByCategory(c.Request.Context(), categoryID)
	if err != nil {
		respondTemplateError(c, err)
		return
	}
	contracts.ResponseSuccess(c, effective)
}

func respondTemplateError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, "unknown error")
		return
	}
	if vErrs, ok := err.(categoryservice.ValidationErrors); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
		return
	}
	if err == gorm.ErrRecordNotFound {
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "template not found")
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "record not found")
		return
	}
	contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
}
