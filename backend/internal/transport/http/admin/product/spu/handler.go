package spu

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP scaffolds for SPU CRUD flows.
type Handler struct {
	service *spuservice.Service
}

// NewHandler constructs a handler wrapper (service may be nil for dry runs).
func NewHandler(service *spuservice.Service) *Handler {
	return &Handler{service: service}
}

// List returns a placeholder payload to unblock frontend wiring.
func (h *Handler) List(c *gin.Context) {
	contracts.ResponseSuccessWithMessage(c, gin.H{
		"items": []any{},
		"meta":  gin.H{"total": 0},
	}, "spu listing scaffold ready")
}

// Create validates payloads before later phases wire persistence logic.
func (h *Handler) Create(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "spu service unavailable")
		return
	}
	var req spuservice.UpsertSPURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	if err := h.service.ValidateUpsertRequest(c.Request.Context(), req); err != nil {
		if vErrs, ok := err.(spuservice.ValidationErrors); ok {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"status": "validated"}, "spu payload accepted")
}

// Get returns a placeholder entry to keep routing stable during initial scaffolding.
func (h *Handler) Get(c *gin.Context) {
	contracts.ResponseSuccessWithMessage(c, gin.H{
		"id":      c.Param("id"),
		"status":  "draft",
		"message": "spu get placeholder",
	}, "spu detail scaffold ready")
}
