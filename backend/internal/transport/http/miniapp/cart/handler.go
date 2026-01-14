package cart

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	cartsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/cart"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *cartsvc.Service
}

func NewHandler(svc *cartsvc.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetCart(c *gin.Context) {
	if h == nil || h.service == nil {
		respondCartError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "购物车服务不可用")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)

	resp, err := h.service.GetCart(c.Request.Context(), tenantUUID, cc.CustomerID)
	if err != nil {
		respondCartError(c, statusForCartError(err), codeForCartError(err), messageForCartError(err))
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) SyncCart(c *gin.Context) {
	if h == nil || h.service == nil {
		respondCartError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "购物车服务不可用")
		return
	}
	var req cartsvc.CartSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondCartError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Strategy) == "" {
		req.Strategy = "max"
	}

	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)

	resp, err := h.service.SyncCart(c.Request.Context(), tenantUUID, cc.CustomerID, req)
	if err != nil {
		respondCartError(c, statusForCartError(err), codeForCartError(err), messageForCartError(err))
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func statusForCartError(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, cartsvc.ErrCustomerRequired):
		return http.StatusUnauthorized
	case errors.Is(err, cartsvc.ErrInvalidQty):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func codeForCartError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, cartsvc.ErrCustomerRequired):
		return contracts.ErrCodeUnauthorized
	case errors.Is(err, cartsvc.ErrInvalidQty):
		return contracts.ErrCodeInvalidRequest
	default:
		return contracts.ErrCodeInvalidRequest
	}
}

func messageForCartError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, cartsvc.ErrCustomerRequired):
		return "未登录或客户信息缺失"
	case errors.Is(err, cartsvc.ErrInvalidQty):
		return "购买数量必须大于 0"
	default:
		return err.Error()
	}
}

func respondCartError(c *gin.Context, status int, code, message string) {
	contracts.ResponseError(c, status, code, message)
}
