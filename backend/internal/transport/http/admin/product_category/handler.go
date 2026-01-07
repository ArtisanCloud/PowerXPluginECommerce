package product_category

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
)

// Handler exposes placeholder endpoints for category domain.
type Handler struct {
	service *categoryservice.Service
}

func NewHandler(service *categoryservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) NotImplemented(c *gin.Context) {
	_ = h
	contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "product category service not implemented")
}
