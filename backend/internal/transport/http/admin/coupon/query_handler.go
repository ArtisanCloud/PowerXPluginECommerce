package coupon

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	service *couponsvc.QueryService
}

func NewQueryHandler(service *couponsvc.QueryService) *QueryHandler {
	return &QueryHandler{service: service}
}

func (h *QueryHandler) ListAssets(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "coupon query service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	out, err := h.service.ListAssets(ctx, tenantUUID, couponsvc.AssetQueryFilter{
		TemplateID: strings.TrimSpace(c.Query("templateId")),
		UserID:     strings.TrimSpace(c.Query("userId")),
		Status:     strings.TrimSpace(c.Query("status")),
		OrderID:    strings.TrimSpace(c.Query("orderId")),
		CouponCode: strings.TrimSpace(c.Query("couponCode")),
		Page:       parsePage(c.Query("page"), 1),
		PageSize:   parsePage(c.Query("pageSize"), 20),
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *QueryHandler) ListUsageLogs(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "coupon query service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	out, err := h.service.ListUsageLogs(ctx, tenantUUID, couponsvc.UsageLogQueryFilter{
		AssetID:    strings.TrimSpace(c.Query("assetId")),
		OrderID:    strings.TrimSpace(c.Query("orderId")),
		Action:     strings.TrimSpace(c.Query("action")),
		CouponCode: strings.TrimSpace(c.Query("couponCode")),
		UserID:     strings.TrimSpace(c.Query("userId")),
		Page:       parsePage(c.Query("page"), 1),
		PageSize:   parsePage(c.Query("pageSize"), 20),
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, out)
}
