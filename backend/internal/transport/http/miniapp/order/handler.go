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
		respondMiniAppError(c, errors.New("order service unavailable"))
		return
	}
	var req ordersvc.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondMiniAppError(c, err)
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	idemKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))

	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.CreateOrder(ctx, tenantUUID, cc.CustomerID, idemKey, req)
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) ListOrders(c *gin.Context) {
	if h == nil || h.service == nil {
		respondMiniAppError(c, errors.New("order service unavailable"))
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListOrders(ctx, tenantUUID, cc.CustomerID, page, pageSize)
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) GetOrder(c *gin.Context) {
	if h == nil || h.service == nil {
		respondMiniAppError(c, errors.New("order service unavailable"))
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	orderID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.GetOrderDetail(ctx, tenantUUID, cc.CustomerID, orderID)
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func httpStatusForOrderError(err error) (status int, code string) {
	if err == nil {
		return http.StatusOK, ""
	}
	switch {
	case errors.Is(err, ordersvc.ErrCustomerRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired),
		errors.Is(err, ordersvc.ErrChannelRequired),
		errors.Is(err, ordersvc.ErrItemsRequired),
		errors.Is(err, ordersvc.ErrInvalidQty),
		errors.Is(err, ordersvc.ErrDuplicateSKU):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
	case errors.Is(err, ordersvc.ErrIdempotencyConflict),
		errors.Is(err, ordersvc.ErrIdempotencyInProgress),
		errors.Is(err, ordersvc.ErrOutOfStock),
		errors.Is(err, ordersvc.ErrSellabilityFailed):
		return http.StatusConflict, contracts.ErrCodeConflict
	case errors.Is(err, ordersvc.ErrOrderNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	default:
		return http.StatusInternalServerError, contracts.ErrCodeInternalError
	}
}

func respondMiniAppError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	status, code := httpStatusForOrderError(err)
	msg := messageForOrderError(err)
	contracts.ResponseError(c, status, code, msg)
}

func messageForOrderError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ordersvc.ErrOrderServiceUnavailable):
		return "订单服务不可用"
	case errors.Is(err, ordersvc.ErrCustomerRequired):
		return "未登录或客户信息缺失"
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired):
		return "缺少幂等键（Idempotency-Key）"
	case errors.Is(err, ordersvc.ErrIdempotencyConflict):
		return "幂等键冲突：请求参数与历史不一致"
	case errors.Is(err, ordersvc.ErrIdempotencyInProgress):
		return "请求处理中，请稍后重试"
	case errors.Is(err, ordersvc.ErrChannelRequired):
		return "channel 必填"
	case errors.Is(err, ordersvc.ErrItemsRequired):
		return "items 不能为空"
	case errors.Is(err, ordersvc.ErrInvalidQty):
		return "购买数量必须大于 0"
	case errors.Is(err, ordersvc.ErrDuplicateSKU):
		return "同一订单中 SKU 不可重复"
	case errors.Is(err, ordersvc.ErrSellabilityFailed):
		return "商品不可售"
	case errors.Is(err, ordersvc.ErrOutOfStock):
		return "库存不足"
	case errors.Is(err, ordersvc.ErrOrderNotFound):
		return "订单不存在"
	default:
		return err.Error()
	}
}

func requestIDFromRequest(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v := strings.TrimSpace(c.GetHeader("X-Request-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.GetHeader("Request-ID")); v != "" {
		return v
	}
	return ""
}
