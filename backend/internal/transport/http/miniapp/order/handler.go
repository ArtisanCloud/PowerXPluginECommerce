package order

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/order"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *ordersvc.Service
}

func NewHandler(svc *ordersvc.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	if h == nil || h.service == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("order service unavailable"))
		return
	}
	var req ordersvc.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondMiniAppError(c, http.StatusBadRequest, err)
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	idemKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))

	resp, err := h.service.CreateOrder(c.Request.Context(), tenantUUID, cc.CustomerID, idemKey, req)
	if err != nil {
		respondMiniAppError(c, httpStatusForOrderError(err), err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) ListOrders(c *gin.Context) {
	if h == nil || h.service == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("order service unavailable"))
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	resp, err := h.service.ListOrders(c.Request.Context(), tenantUUID, cc.CustomerID, page, pageSize)
	if err != nil {
		respondMiniAppError(c, httpStatusForOrderError(err), err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) GetOrder(c *gin.Context) {
	if h == nil || h.service == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("order service unavailable"))
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	orderID := strings.TrimSpace(c.Param("id"))
	resp, err := h.service.GetOrderDetail(c.Request.Context(), tenantUUID, cc.CustomerID, orderID)
	if err != nil {
		respondMiniAppError(c, httpStatusForOrderError(err), err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func httpStatusForOrderError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errors.Is(err, ordersvc.ErrCustomerRequired):
		return http.StatusUnauthorized
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired),
		errors.Is(err, ordersvc.ErrChannelRequired),
		errors.Is(err, ordersvc.ErrItemsRequired),
		errors.Is(err, ordersvc.ErrInvalidQty),
		errors.Is(err, ordersvc.ErrDuplicateSKU):
		return http.StatusBadRequest
	case errors.Is(err, ordersvc.ErrIdempotencyConflict),
		errors.Is(err, ordersvc.ErrIdempotencyInProgress),
		errors.Is(err, ordersvc.ErrOutOfStock),
		errors.Is(err, ordersvc.ErrSellabilityFailed):
		return http.StatusConflict
	case errors.Is(err, ordersvc.ErrOrderNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func respondMiniAppError(c *gin.Context, status int, err error) {
	if c == nil {
		return
	}
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}
