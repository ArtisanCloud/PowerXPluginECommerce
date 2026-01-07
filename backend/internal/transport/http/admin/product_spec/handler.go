package product_spec

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	specservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_spec"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *specservice.Service
}

func NewHandler(service *specservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "spec service unavailable", nil)
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		contracts.ResponseBadRequest(c, "spu id is required")
		return
	}
	items, err := h.service.List(c.Request.Context(), spuID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"groups": items})
}

func (h *Handler) Replace(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "spec service unavailable", nil)
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		contracts.ResponseBadRequest(c, "spu id is required")
		return
	}
	var req specservice.ReplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	items, err := h.service.Replace(c.Request.Context(), spuID, req)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"groups": items}, "spec updated")
}

