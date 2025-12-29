package product_sku

import (
	"net/http"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// ChannelHandler exposes mapping + publish endpoints.
type ChannelHandler struct {
	service *productskuservice.Service
}

func NewChannelHandler(service *productskuservice.Service) *ChannelHandler {
	return &ChannelHandler{service: service}
}

func (h *ChannelHandler) List(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	mappings, err := h.service.ListChannelMappings(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mappings)
}

func (h *ChannelHandler) Upsert(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	var req productskuservice.ChannelMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	mapping, err := h.service.UpsertChannelMapping(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapping)
}

func (h *ChannelHandler) Publish(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	var req productskuservice.ChannelPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	resp, err := h.service.PublishChannelMapping(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}
