package product_sku

import (
	"net/http"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// BarcodeHandler exposes barcode generation / validation endpoints.
type BarcodeHandler struct {
	service *productskuservice.Service
}

func NewBarcodeHandler(service *productskuservice.Service) *BarcodeHandler {
	return &BarcodeHandler{service: service}
}

func (h *BarcodeHandler) Generate(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	skuID := c.Param("id")
	var req productskuservice.BarcodeGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	result, err := h.service.GenerateBarcodes(c.Request.Context(), skuID, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
