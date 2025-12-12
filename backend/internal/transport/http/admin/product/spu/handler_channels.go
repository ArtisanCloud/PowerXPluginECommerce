package spu

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// ChannelHandler exposes HTTP handlers for channel visibility operations.
type ChannelHandler struct {
	service *spuservice.ChannelService
}

// NewChannelHandler constructs a handler wrapper for channel operations.
func NewChannelHandler(service *spuservice.ChannelService) *ChannelHandler {
	return &ChannelHandler{service: service}
}

// List returns channels for a single SPU.
func (h *ChannelHandler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	items, err := h.service.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": items})
}

// Upsert creates or updates a single channel visibility entry.
func (h *ChannelHandler) Upsert(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var req spuservice.ChannelVisibilityInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	entry, err := h.service.Upsert(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, entry, "channel upserted")
}

// Delete removes the specific channel configuration from an SPU.
func (h *ChannelHandler) Delete(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	if err := h.service.Remove(c.Request.Context(), c.Param("id"), c.Param("channel")); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{}, "channel removed")
}
