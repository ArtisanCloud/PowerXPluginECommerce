package coupon

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	service *couponsvc.TemplateService
}

func NewTemplateHandler(service *couponsvc.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

func (h *TemplateHandler) List(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "coupon template service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	res, err := h.service.List(ctx, tenantUUID, couponsvc.TemplateListFilter{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     parsePage(c.Query("page"), 1),
		PageSize: parsePage(c.Query("pageSize"), 20),
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, res)
}

func (h *TemplateHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "coupon template service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	var req couponsvc.TemplateCreateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	req.TenantUUID = tenantUUID
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	out, err := h.service.Create(ctx, req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *TemplateHandler) Update(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "coupon template service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if _, ok := requireAdminUser(c); !ok {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "id is required")
		return
	}
	var req couponsvc.TemplateUpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	out, err := h.service.Update(ctx, tenantUUID, id, req)
	if err != nil {
		if errors.Is(err, couponsvc.ErrCouponTemplateNotFound) {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, err.Error())
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, out)
}
