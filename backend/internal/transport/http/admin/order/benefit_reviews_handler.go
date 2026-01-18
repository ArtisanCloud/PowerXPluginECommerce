package order

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/order"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type BenefitReviewHandler struct {
	service *ordersvc.BenefitReviewService
}

func NewBenefitReviewHandler(svc *ordersvc.BenefitReviewService) *BenefitReviewHandler {
	return &BenefitReviewHandler{service: svc}
}

func (h *BenefitReviewHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "benefit review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if tc, ok := authx.GetTenantContext(c); ok {
		ctx = authx.ContextWithTenantContext(ctx, tc)
	}
	resp, err := h.service.ListReviews(ctx, tenantUUID, adminID, orderID)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *BenefitReviewHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "benefit review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	var req ordersvc.BenefitReviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if tc, ok := authx.GetTenantContext(c); ok {
		ctx = authx.ContextWithTenantContext(ctx, tc)
	}
	resp, err := h.service.CreateReview(ctx, tenantUUID, adminID, orderID, req)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *BenefitReviewHandler) Approve(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "benefit review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var req ordersvc.BenefitReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if tc, ok := authx.GetTenantContext(c); ok {
		ctx = authx.ContextWithTenantContext(ctx, tc)
	}
	resp, err := h.service.ApproveReviews(ctx, tenantUUID, adminID, req.ReviewIDs, req)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *BenefitReviewHandler) Reject(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "benefit review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var req ordersvc.BenefitReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if tc, ok := authx.GetTenantContext(c); ok {
		ctx = authx.ContextWithTenantContext(ctx, tc)
	}
	resp, err := h.service.RejectReviews(ctx, tenantUUID, adminID, req.ReviewIDs, req)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *BenefitReviewHandler) SearchCodes(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "benefit review service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	benefitType := strings.TrimSpace(c.Query("type"))
	keyword := strings.TrimSpace(c.Query("q"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	if tc, ok := authx.GetTenantContext(c); ok {
		ctx = authx.ContextWithTenantContext(ctx, tc)
	}
	resp, err := h.service.SearchBenefitCodes(ctx, tenantUUID, adminID, benefitType, keyword)
	if err != nil {
		respondAdminOrderError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}
