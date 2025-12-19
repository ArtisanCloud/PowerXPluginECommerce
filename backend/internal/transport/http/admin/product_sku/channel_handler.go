package product_sku

import (
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
	respondStub(c, "SKU channel list not implemented")
}

func (h *ChannelHandler) Upsert(c *gin.Context) {
	respondStub(c, "SKU channel upsert not implemented")
}

func (h *ChannelHandler) Publish(c *gin.Context) {
	respondStub(c, "SKU channel publish not implemented")
}
