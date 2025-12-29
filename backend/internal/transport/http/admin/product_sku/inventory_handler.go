package product_sku

import (
	"net/http"

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
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	snapshot, err := h.service.SnapshotInventory(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}
