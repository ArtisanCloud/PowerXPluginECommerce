package product_sku

import (
	"errors"
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// BulkHandler wires `/products/skus/bulk-tasks` endpoints.
type BulkHandler struct {
	service *productskuservice.Service
}

type bulkApprovalRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

func NewBulkHandler(service *productskuservice.Service) *BulkHandler {
	return &BulkHandler{service: service}
}

func (h *BulkHandler) Submit(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var req productskuservice.BulkAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	if tc, ok := middleware.GetTenantContext(c); ok {
		req.TenantUUID = tc.TenantUUID
		req.RequestedBy = tc.Actor().ID()
		req.RequestedByRoles = tc.Roles
	}
	resp, err := h.service.SubmitBulkTask(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}

func (h *BulkHandler) Get(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	resp, err := h.service.GetBulkTask(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BulkHandler) DecideApproval(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var req bulkApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	actor := "system"
	if tc, ok := middleware.GetTenantContext(c); ok {
		actor = tc.Actor().ID()
	}
	resp, err := h.service.DecideBulkTaskApproval(
		c.Request.Context(),
		c.Param("taskId"),
		productskuservice.BulkTaskApprovalDecision{
			Decision: req.Decision,
			Note:     req.Note,
		},
		actor,
	)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BulkHandler) Retry(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	resp, err := h.service.RetryBulkTask(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}
