package order

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/order"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
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
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "order service unavailable")
		return
	}

	var req ordersvc.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}

	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return
	}
	adminID := strconv.FormatInt(tc.UserID, 10)

	idemKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	resp, err := h.service.CreateOrder(c.Request.Context(), tenantUUID, adminID, idemKey, req)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) ListOrders(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "order service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	status := strings.TrimSpace(c.Query("status"))
	customerID := strings.TrimSpace(c.Query("customerId"))

	resp, err := h.service.ListOrders(c.Request.Context(), tenantUUID, orderrepo.OrderListFilter{
		CustomerID: customerID,
		Status:     status,
	}, page, pageSize)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) GetOrder(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "order service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	resp, err := h.service.GetOrderDetail(c.Request.Context(), tenantUUID, orderID)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "order service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))

	var req ordersvc.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}

	resp, err := h.service.CancelOrder(c.Request.Context(), tenantUUID, adminID, orderID, req.Reason)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func requireAdminUser(c *gin.Context) (string, bool) {
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return "", false
	}
	return strconv.FormatInt(tc.UserID, 10), true
}
