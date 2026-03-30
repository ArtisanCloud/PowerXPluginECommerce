package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListOpsAutomationPolicies(c *gin.Context) {
	if h == nil || h.opsAutomationSvc == nil {
		contracts.ResponseServiceUnavailable(c, "ops automation service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	var enabled *bool
	if raw := strings.TrimSpace(c.Query("enabled")); raw != "" {
		v := strings.EqualFold(raw, "1") || strings.EqualFold(raw, "true")
		enabled = &v
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.opsAutomationSvc.ListPolicies(c.Request.Context(), tenantUUID, logisticssvc.OpsAutomationPolicyQuery{
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		Enabled:   enabled,
		Limit:     limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertOpsAutomationPolicy(c *gin.Context) {
	if h == nil || h.opsAutomationSvc == nil {
		contracts.ResponseServiceUnavailable(c, "ops automation service unavailable", nil)
		return
	}
	var payload upsertOpsAutomationPolicyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.opsAutomationSvc.UpsertPolicy(c.Request.Context(), tenantUUID, logisticssvc.UpsertOpsAutomationPolicyRequest{
		ID:                     strings.TrimSpace(payload.ID),
		Name:                   strings.TrimSpace(payload.Name),
		CarrierID:              strings.TrimSpace(payload.CarrierID),
		RetryStrategy:          payload.RetryStrategy,
		CircuitBreakerStrategy: payload.CircuitBreakerStrategy,
		SuppressionRule:        payload.SuppressionRule,
		EscalationChain:        payload.EscalationChain,
		Enabled:                payload.Enabled,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) EvaluateOpsAutomation(c *gin.Context) {
	if h == nil || h.opsAutomationSvc == nil {
		contracts.ResponseServiceUnavailable(c, "ops automation service unavailable", nil)
		return
	}
	var payload evaluateOpsAutomationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.opsAutomationSvc.Evaluate(c.Request.Context(), tenantUUID, logisticssvc.EvaluateOpsAutomationRequest{
		CarrierID:     strings.TrimSpace(payload.CarrierID),
		TriggerSource: strings.TrimSpace(payload.TriggerSource),
		Limit:         payload.Limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListOpsAutomationRuns(c *gin.Context) {
	if h == nil || h.opsAutomationSvc == nil {
		contracts.ResponseServiceUnavailable(c, "ops automation service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.opsAutomationSvc.ListRuns(c.Request.Context(), tenantUUID, logisticssvc.OpsAutomationRunQuery{
		PolicyID:  strings.TrimSpace(c.Query("policy_id")),
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		Status:    strings.TrimSpace(c.Query("status")),
		Limit:     limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) TakeoverOpsAutomationRun(c *gin.Context) {
	if h == nil || h.opsAutomationSvc == nil {
		contracts.ResponseServiceUnavailable(c, "ops automation service unavailable", nil)
		return
	}
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		contracts.ResponseBadRequest(c, "run id is required")
		return
	}
	var payload takeoverOpsAutomationRunRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.opsAutomationSvc.Takeover(c.Request.Context(), tenantUUID, runID, logisticssvc.OpsAutomationTakeoverRequest{
		Action:     strings.TrimSpace(payload.Action),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
