package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// InventoryHandler provides read-only SKU inventory snapshots.
type InventoryHandler struct {
	service *productskuservice.Service
}

func NewInventoryHandler(service *productskuservice.Service) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) Get(c *gin.Context) {
	respondStub(c, "SKU inventory view not implemented")
}
