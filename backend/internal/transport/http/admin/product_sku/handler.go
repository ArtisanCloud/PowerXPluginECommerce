package product_sku

import (
	"net/http"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

func respondStub(c *gin.Context, message string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": message,
	})
}

// Handler wires basic CRUD endpoints for SKU resources.
type Handler struct {
	service *productskuservice.Service
}

func NewHandler(service *productskuservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	respondStub(c, "SKU listing not implemented")
}

func (h *Handler) Upsert(c *gin.Context) {
	respondStub(c, "SKU create/update not implemented")
}

func (h *Handler) Update(c *gin.Context) {
	respondStub(c, "SKU update not implemented")
}

func (h *Handler) Delete(c *gin.Context) {
	respondStub(c, "SKU delete not implemented")
}
