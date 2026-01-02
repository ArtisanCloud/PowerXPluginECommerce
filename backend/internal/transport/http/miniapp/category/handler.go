package category

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
)

// Handler exposes read-only category endpoints for mini-app clients.
type Handler struct {
	service *categoryservice.Service
}

func NewHandler(service *categoryservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Tree(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	nodes, err := h.service.MiniAppTree(c.Request.Context())
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": nodes})
}
