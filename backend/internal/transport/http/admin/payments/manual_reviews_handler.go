package payments

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type ManualReviewHandler struct {
	service *paymentsvc.ManualReviewService
}

func NewManualReviewHandler(svc *paymentsvc.ManualReviewService) *ManualReviewHandler {
	return &ManualReviewHandler{service: svc}
}

func (h *ManualReviewHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "manual review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	orderID := strings.TrimSpace(c.Query("orderId"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListReviews(ctx, tenantUUID, adminID, orderID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ManualReviewHandler) ListLogs(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "manual review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return
	}
	adminID := strconv.FormatInt(tc.UserID, 10)
	orderID := strings.TrimSpace(c.Query("orderId"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	ctx = authx.ContextWithTenantContext(ctx, tc)
	resp, err := h.service.ListReviewLogs(ctx, tenantUUID, adminID, orderID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ManualReviewHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "manual review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return
	}
	adminID := strconv.FormatInt(tc.UserID, 10)
	var req paymentsvc.ManualPaymentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	ctx = authx.ContextWithTenantContext(ctx, tc)
	resp, err := h.service.CreateReview(ctx, tenantUUID, adminID, req)
	if err != nil {
		if errors.Is(err, paymentsvc.ErrManualReviewOrderNotPayable) {
			contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseCreated(c, resp)
}

func (h *ManualReviewHandler) Approve(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "manual review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return
	}
	adminID := strconv.FormatInt(tc.UserID, 10)
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid review id")
		return
	}
	var req paymentsvc.ManualPaymentReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	ctx = authx.ContextWithTenantContext(ctx, tc)
	resp, err := h.service.ApproveReview(ctx, tenantUUID, adminID, id, req)
	if err != nil {
		if errors.Is(err, paymentsvc.ErrManualReviewNotFound) {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, err.Error())
			return
		}
		if errors.Is(err, paymentsvc.ErrManualReviewInvalidStatus) || errors.Is(err, paymentsvc.ErrManualReviewSameOperator) {
			contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
		if errors.Is(err, paymentsvc.ErrManualReviewOrderNotPayable) {
			contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *ManualReviewHandler) Reject(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "manual review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return
	}
	adminID := strconv.FormatInt(tc.UserID, 10)
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid review id")
		return
	}
	var req paymentsvc.ManualPaymentReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	ctx = authx.ContextWithTenantContext(ctx, tc)
	resp, err := h.service.RejectReview(ctx, tenantUUID, adminID, id, req)
	if err != nil {
		if errors.Is(err, paymentsvc.ErrManualReviewNotFound) {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, err.Error())
			return
		}
		if errors.Is(err, paymentsvc.ErrManualReviewInvalidStatus) || errors.Is(err, paymentsvc.ErrManualReviewSameOperator) || errors.Is(err, paymentsvc.ErrManualReviewReasonRequired) {
			contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
			return
		}
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
