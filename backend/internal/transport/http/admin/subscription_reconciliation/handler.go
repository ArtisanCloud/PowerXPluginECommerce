package subscription_reconciliation

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	SubscriptionReconciliationSvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/subscription_reconciliation"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *SubscriptionReconciliationSvc.Service
}

func NewHandler(service *SubscriptionReconciliationSvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateBatch(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	var payload createBatchRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	payload = payload.normalize()
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.service.CreateBatch(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.CreateBatchInput{
		BillingCycle: payload.BillingCycle,
		RunType:      payload.RunType,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListBatches(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.service.ListBatches(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.BatchListQuery{
		BillingCycle: strings.TrimSpace(c.Query("billingCycle")),
		Status:       strings.TrimSpace(c.Query("status")),
		Limit:        limit,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ListBatchDeltas(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	batchID := strings.TrimSpace(c.Param("id"))
	if batchID == "" {
		contracts.ResponseBadRequest(c, "batch id is required")
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.service.ListDeltas(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.DeltaListQuery{
		BatchID:   batchID,
		DeltaType: strings.TrimSpace(c.Query("deltaType")),
		RiskLevel: strings.TrimSpace(c.Query("riskLevel")),
		Limit:     limit,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateDeltaTask(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	deltaID := strings.TrimSpace(c.Param("id"))
	if deltaID == "" {
		contracts.ResponseBadRequest(c, "delta id is required")
		return
	}
	var payload createDeltaTaskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	payload = payload.normalize()
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.service.CreateDeltaTask(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.CreateDeltaTaskInput{
		DeltaID:  deltaID,
		Assignee: payload.Assignee,
		SLALevel: payload.SLALevel,
		Note:     payload.Note,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) AdjustDelta(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	deltaID := strings.TrimSpace(c.Param("id"))
	if deltaID == "" {
		contracts.ResponseBadRequest(c, "delta id is required")
		return
	}
	var payload adjustDeltaRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	payload = payload.normalize()
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.service.AdjustDelta(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.AdjustDeltaInput{
		DeltaID:             deltaID,
		ExpectedAmountMinor: payload.ExpectedAmountMinor,
		ActualAmountMinor:   payload.ActualAmountMinor,
		ReasonCode:          payload.ReasonCode,
		Note:                payload.Note,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) CloseTask(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	var payload closeTaskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	payload = payload.normalize()
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.service.CloseTask(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.CloseTaskInput{
		TaskID:         taskID,
		Resolution:     payload.Resolution,
		ResolutionNote: payload.ResolutionNote,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) Dashboard(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.Dashboard(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.DashboardQuery{
		From: strings.TrimSpace(c.Query("from")),
		To:   strings.TrimSpace(c.Query("to")),
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, SubscriptionReconciliationSvc.ErrServiceUnavailable):
		contracts.ResponseServiceUnavailable(c, err.Error(), nil)
	case errors.Is(err, SubscriptionReconciliationSvc.ErrNotImplemented):
		contracts.ResponseError(c, http.StatusNotImplemented, contracts.ErrCodeReconciliationUnsupportedAction, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrInvalidBillingCycle):
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeReconciliationInvalidBillingCycle, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrInvalidRunType),
		errors.Is(err, SubscriptionReconciliationSvc.ErrInvalidSLALevel),
		errors.Is(err, SubscriptionReconciliationSvc.ErrInvalidResolution):
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeReconciliationUnsupportedAction, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrBatchNotFound):
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeReconciliationBatchNotFound, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrDeltaNotFound):
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeReconciliationDeltaNotFound, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrTaskAlreadyOpen):
		contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeReconciliationTaskAlreadyOpen, err.Error())
	case errors.Is(err, SubscriptionReconciliationSvc.ErrTaskNotFound):
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeReconciliationDeltaNotFound, err.Error())
	default:
		contracts.ResponseInternalError(c, err)
	}
}
