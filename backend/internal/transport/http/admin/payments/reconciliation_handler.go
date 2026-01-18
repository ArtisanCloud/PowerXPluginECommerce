package payments

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type ReconciliationHandler struct {
	service *paymentsvc.ReconciliationService
}

func NewReconciliationHandler(svc *paymentsvc.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{service: svc}
}

func (h *ReconciliationHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "reconciliation service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListReconciliations(ctx, tenantUUID, adminID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ReconciliationHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "reconciliation service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var req paymentsvc.CreateReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid reconciliation payload")
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.CreateReconciliation(ctx, tenantUUID, adminID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, paymentsvc.ErrReconciliationInvalidPeriod) || errors.Is(err, paymentsvc.ErrReconciliationItemsRequired) {
			status = http.StatusBadRequest
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ReconciliationHandler) ListItems(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "reconciliation service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	idRaw := strings.TrimSpace(c.Param("id"))
	reconciliationID, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid reconciliation id")
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListReconciliationItems(ctx, tenantUUID, adminID, reconciliationID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ReconciliationHandler) ResolveItem(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "reconciliation service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	idRaw := strings.TrimSpace(c.Param("id"))
	reconciliationID, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid reconciliation id")
		return
	}
	itemRaw := strings.TrimSpace(c.Param("itemId"))
	itemID, err := strconv.ParseUint(itemRaw, 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid reconciliation item id")
		return
	}
	var req paymentsvc.ResolveReconciliationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid reconciliation resolution payload")
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ResolveReconciliationItem(ctx, tenantUUID, adminID, reconciliationID, itemID, req)
	if err != nil {
		if errors.Is(err, paymentsvc.ErrReconciliationItemNotFound) {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, err.Error())
			return
		}
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
