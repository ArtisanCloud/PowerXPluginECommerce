package spu

import (
	"net/http"
	"strconv"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler exposes HTTP scaffolds for SPU CRUD flows.
type Handler struct {
	service *spuservice.Service
}

// NewHandler constructs a handler wrapper (service may be nil for dry runs).
func NewHandler(service *spuservice.Service) *Handler {
	return &Handler{service: service}
}

// List returns SPU summaries.
func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	filters := spuservice.ListFilters{
		Keyword:            c.Query("keyword"),
		Status:             c.Query("status"),
		Type:               c.Query("type"),
		CategoryID:         c.Query("categoryId"),
		CategoryPathPrefix: c.Query("categoryPathPrefix"),
		Page:               toInt(c.Query("page"), 1),
		PageSize:           toInt(c.Query("pageSize"), 20),
	}
	result, err := h.service.List(c.Request.Context(), filters)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"items": result.Items,
		"meta":  gin.H{"total": result.Total, "page": result.Page, "pageSize": result.PageSize},
	})
}

// Create validates payloads and persists a draft SPU.
func (h *Handler) Create(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.UpsertSPURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	detail, err := h.service.CreateDraft(c.Request.Context(), req)
	if err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu draft created")
}

// Update edits an existing draft SPU.
func (h *Handler) Update(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.UpsertSPURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	detail, err := h.service.UpdateDraft(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu draft updated")
}

// Submit sends a draft into review state.
func (h *Handler) Submit(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	req := spuservice.SubmitRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
	}
	detail, err := h.service.Submit(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu submitted for review")
}

// Publish transitions a reviewing SPU to published and triggers channel tasks.
func (h *Handler) Publish(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	detail, err := h.service.Publish(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu publish request accepted")
}

// Withdraw downlines specific channels for an SPU.
func (h *Handler) Withdraw(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	detail, err := h.service.Withdraw(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu withdraw request accepted")
}

// Delete soft deletes a draft/offboarded SPU.
func (h *Handler) Delete(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	detail, err := h.service.Delete(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu deleted")
}

// Get returns SPU detail by id.
func (h *Handler) Get(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	detail, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "spu not found")
			return
		}
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, detail)
}

// Revise creates a new draft version from current published SPU (Plan A).
func (h *Handler) Revise(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	req := spuservice.ReviseRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
	}
	detail, err := h.service.Revise(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "spu revised to draft")
}

func toInt(val string, fallback int) int {
	if val == "" {
		return fallback
	}
	if n, err := strconv.Atoi(val); err == nil && n > 0 {
		return n
	}
	return fallback
}
