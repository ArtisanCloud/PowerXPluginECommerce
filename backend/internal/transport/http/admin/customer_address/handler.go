package customer_address

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	addresssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/customer_address"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *addresssvc.Service
}

func NewHandler(svc *addresssvc.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "address service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	customerID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListAddresses(ctx, tenantUUID, adminID, customerID)
	if err != nil {
		respondAdminAddressError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "address service unavailable")
		return
	}
	var req addresssvc.AddressUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	customerID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.CreateAddress(ctx, tenantUUID, adminID, customerID, req)
	if err != nil {
		respondAdminAddressError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Update(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "address service unavailable")
		return
	}
	var req addresssvc.AddressUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	addrID := strings.TrimSpace(c.Param("addressId"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	customerID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.UpdateAddress(ctx, tenantUUID, adminID, customerID, addrID, req)
	if err != nil {
		respondAdminAddressError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "address service unavailable")
		return
	}
	addrID := strings.TrimSpace(c.Param("addressId"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	customerID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if err := h.service.DeleteAddress(ctx, tenantUUID, adminID, customerID, addrID); err != nil {
		respondAdminAddressError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"success": true})
}

func (h *Handler) SetDefault(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "address service unavailable")
		return
	}
	addrID := strings.TrimSpace(c.Param("addressId"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	customerID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.SetDefault(ctx, tenantUUID, adminID, customerID, addrID)
	if err != nil {
		respondAdminAddressError(c, err)
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

func respondAdminAddressError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	status, code := httpStatusForAddressError(err)
	msg := messageForAddressError(err)
	contracts.ResponseError(c, status, code, msg)
}

func httpStatusForAddressError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	switch {
	case errors.Is(err, addresssvc.ErrAdminRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized
	case errors.Is(err, addresssvc.ErrCustomerRequired),
		errors.Is(err, addresssvc.ErrInvalidAddress):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
	case errors.Is(err, addresssvc.ErrAddressNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	default:
		return http.StatusInternalServerError, contracts.ErrCodeInternalError
	}
}

func messageForAddressError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, addresssvc.ErrAdminRequired):
		return "未授权"
	case errors.Is(err, addresssvc.ErrCustomerRequired):
		return "customerId 必填"
	case errors.Is(err, addresssvc.ErrInvalidAddress):
		return "收货地址不完整"
	case errors.Is(err, addresssvc.ErrAddressNotFound):
		return "收货地址不存在"
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
