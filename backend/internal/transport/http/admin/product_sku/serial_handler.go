package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// SerialHandler controls serial / batch level CRUD endpoints.
type SerialHandler struct {
	service *productskuservice.Service
}

func NewSerialHandler(service *productskuservice.Service) *SerialHandler {
	return &SerialHandler{service: service}
}

func (h *SerialHandler) List(c *gin.Context) {
	respondStub(c, "SKU serial list not implemented")
}

func (h *SerialHandler) Create(c *gin.Context) {
	respondStub(c, "SKU serial create not implemented")
}
