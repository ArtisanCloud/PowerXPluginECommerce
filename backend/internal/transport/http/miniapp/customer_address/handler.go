package customer_address

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	addresssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/customer_address"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
)

type Handler struct {
	service *addresssvc.Service
}

func NewHandler(svc *addresssvc.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		respondError(c, errors.New("address service unavailable"))
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListAddresses(ctx, tenantUUID, cc.CustomerID)
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		respondError(c, errors.New("address service unavailable"))
		return
	}
	var req addresssvc.AddressUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.CreateAddress(ctx, tenantUUID, cc.CustomerID, req)
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Update(c *gin.Context) {
	if h == nil || h.service == nil {
		respondError(c, errors.New("address service unavailable"))
		return
	}
	var req addresssvc.AddressUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	addrID := strings.TrimSpace(c.Param("id"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.UpdateAddress(ctx, tenantUUID, cc.CustomerID, addrID, req)
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	if h == nil || h.service == nil {
		respondError(c, errors.New("address service unavailable"))
		return
	}
	addrID := strings.TrimSpace(c.Param("id"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if err := h.service.DeleteAddress(ctx, tenantUUID, cc.CustomerID, addrID); err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"success": true})
}

func (h *Handler) SetDefault(c *gin.Context) {
	if h == nil || h.service == nil {
		respondError(c, errors.New("address service unavailable"))
		return
	}
	addrID := strings.TrimSpace(c.Param("id"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.SetDefault(ctx, tenantUUID, cc.CustomerID, addrID)
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func respondError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	logger.WithError(err).WithFields(logger.Fields{
		"request_id": requestIDFromRequest(c),
		"path":       c.FullPath(),
		"method":     c.Request.Method,
	}).Error("mini-app address request failed")
	status, code := httpStatusForError(err)
	contracts.ResponseError(c, status, code, messageForError(err))
}

func httpStatusForError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	switch {
	case isUndefinedTableError(err):
		return http.StatusServiceUnavailable, contracts.ErrCodeInternalError
	case errors.Is(err, addresssvc.ErrCustomerRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized
	case errors.Is(err, addresssvc.ErrInvalidAddress):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
	case errors.Is(err, addresssvc.ErrAddressNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	default:
		return http.StatusInternalServerError, contracts.ErrCodeInternalError
	}
}

func messageForError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case isUndefinedTableError(err):
		return "数据库未初始化/未迁移：请先执行 make migrate 并重启后端"
	case errors.Is(err, addresssvc.ErrCustomerRequired):
		return "未登录或客户信息缺失"
	case errors.Is(err, addresssvc.ErrInvalidAddress):
		return "收货地址不完整"
	case errors.Is(err, addresssvc.ErrAddressNotFound):
		return "收货地址不存在"
	default:
		return err.Error()
	}
}

func isUndefinedTableError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr == nil {
		return false
	}
	return strings.TrimSpace(pgErr.Code) == "42P01"
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
