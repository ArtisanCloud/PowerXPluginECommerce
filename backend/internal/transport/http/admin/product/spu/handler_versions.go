package spu

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VersionHandler exposes HTTP endpoints for version timelines and approvals.
type VersionHandler struct {
	service *spuservice.VersionService
}

// NewVersionHandler constructs a handler with the provided service.
func NewVersionHandler(service *spuservice.VersionService) *VersionHandler {
	return &VersionHandler{service: service}
}

// List returns version summaries for an SPU.
func (h *VersionHandler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "version service unavailable", nil)
		return
	}
	filters := spuservice.VersionListFilters{
		Status:   c.Query("status"),
		Page:     toInt(c.Query("page"), 1),
		PageSize: toInt(c.Query("pageSize"), 20),
	}
	result, err := h.service.List(c.Request.Context(), c.Param("id"), filters)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}

// Get returns a specific version detail with diff and approvals.
func (h *VersionHandler) Get(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "version service unavailable", nil)
		return
	}
	detail, err := h.service.Get(c.Request.Context(), c.Param("id"), c.Param("versionId"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			contracts.ResponseNotFound(c, "version not found")
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, detail)
}

// Approve advances approval chain.
func (h *VersionHandler) Approve(c *gin.Context) {
	h.processApproval(c, true)
}

// Reject marks the version as rejected.
func (h *VersionHandler) Reject(c *gin.Context) {
	h.processApproval(c, false)
}

func (h *VersionHandler) processApproval(c *gin.Context, approve bool) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "version service unavailable", nil)
		return
	}
	var req spuservice.ApprovalActionRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseBadRequest(c, err.Error())
			return
		}
	}
	var (
		detail *spuservice.VersionDetail
		err    error
	)
	if approve {
		detail, err = h.service.Approve(c.Request.Context(), c.Param("id"), c.Param("versionId"), req)
	} else {
		detail, err = h.service.Reject(c.Request.Context(), c.Param("id"), c.Param("versionId"), req)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			contracts.ResponseNotFound(c, "version not found")
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	message := "version rejected"
	if approve {
		message = "approval recorded"
	}
	contracts.ResponseSuccessWithMessage(c, detail, message)
}

// Rollback recreates a draft from a historical version.
func (h *VersionHandler) Rollback(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "version service unavailable", nil)
		return
	}
	var req spuservice.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	if req.TargetVersionID == "" {
		req.TargetVersionID = c.Param("versionId")
	}
	detail, err := h.service.Rollback(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, detail, "rollback draft created")
}
