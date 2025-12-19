package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// BulkHandler wires `/products/skus/bulk-tasks` endpoints.
type BulkHandler struct {
	service *productskuservice.Service
}

func NewBulkHandler(service *productskuservice.Service) *BulkHandler {
	return &BulkHandler{service: service}
}

func (h *BulkHandler) Submit(c *gin.Context) {
	respondStub(c, "SKU bulk task submission not implemented")
}

func (h *BulkHandler) Get(c *gin.Context) {
	respondStub(c, "SKU bulk task query not implemented")
}
