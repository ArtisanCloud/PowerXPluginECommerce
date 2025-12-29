package product_sku

import (
	"errors"
	"net/http"

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
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var req productskuservice.SkuGeneratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	resp, err := h.service.GenerateSkusFromSPU(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
