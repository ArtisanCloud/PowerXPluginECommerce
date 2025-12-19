package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// ImportExportHandler wires SKU import/export placeholder endpoints.
type ImportExportHandler struct {
	service *productskuservice.Service
}

func NewImportExportHandler(service *productskuservice.Service) *ImportExportHandler {
	return &ImportExportHandler{service: service}
}

func (h *ImportExportHandler) Import(c *gin.Context) {
	respondStub(c, "SKU import not implemented")
}

func (h *ImportExportHandler) Export(c *gin.Context) {
	respondStub(c, "SKU export not implemented")
}
