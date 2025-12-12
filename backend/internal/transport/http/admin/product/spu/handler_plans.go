package spu

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// PlanHandler exposes subscription plan CRUD endpoints.
type PlanHandler struct {
	service *spuservice.SubscriptionPlanService
}

// NewPlanHandler constructs a plan handler with optional service.
func NewPlanHandler(service *spuservice.SubscriptionPlanService) *PlanHandler {
	return &PlanHandler{service: service}
}

// List returns subscription plans bound to an SPU.
func (h *PlanHandler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "subscription plan service unavailable")
		return
	}
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": items})
}

// Create inserts a new subscription plan.
func (h *PlanHandler) Create(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "subscription plan service unavailable")
		return
	}
	var req spuservice.SubscriptionPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	entry, err := h.service.Create(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, entry, "subscription plan created")
}

// Update edits an existing plan.
func (h *PlanHandler) Update(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "subscription plan service unavailable")
		return
	}
	var req spuservice.SubscriptionPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	entry, err := h.service.Update(c.Request.Context(), c.Param("id"), c.Param("planId"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, entry, "subscription plan updated")
}

// Delete archives the plan.
func (h *PlanHandler) Delete(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "subscription plan service unavailable")
		return
	}
	if err := h.service.Archive(c.Request.Context(), c.Param("id"), c.Param("planId")); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{}, "subscription plan archived")
}
