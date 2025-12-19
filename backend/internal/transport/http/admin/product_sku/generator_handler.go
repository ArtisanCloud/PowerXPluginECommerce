package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// GeneratorHandler hosts the `/spus/:spuId/skus/generate` endpoint.
type GeneratorHandler struct {
	service *productskuservice.Service
}

func NewGeneratorHandler(service *productskuservice.Service) *GeneratorHandler {
	return &GeneratorHandler{service: service}
}

func (h *GeneratorHandler) Generate(c *gin.Context) {
	respondStub(c, "SKU generator not implemented")
}
