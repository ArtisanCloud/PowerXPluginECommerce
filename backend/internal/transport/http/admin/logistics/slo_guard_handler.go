package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListSLOGuardPolicies(c *gin.Context) {
	if h == nil || h.sloGuardSvc == nil {
		contracts.ResponseServiceUnavailable(c, "slo guard service unavailable", nil)
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
	rows, err := h.sloGuardSvc.ListPolicies(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("carrier_id")), enabled, limit)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertSLOGuardPolicy(c *gin.Context) {
	if h == nil || h.sloGuardSvc == nil {
		contracts.ResponseServiceUnavailable(c, "slo guard service unavailable", nil)
		return
	}
	var payload upsertSLOGuardPolicyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sloGuardSvc.UpsertPolicy(c.Request.Context(), tenantUUID, logisticssvc.UpsertSLOGuardPolicyRequest{
		ID:                strings.TrimSpace(payload.ID),
		Name:              strings.TrimSpace(payload.Name),
		CarrierID:         strings.TrimSpace(payload.CarrierID),
		WindowHours:       payload.WindowHours,
		MinSuccessRate:    payload.MinSuccessRate,
		MaxP95LatencyMS:   payload.MaxP95LatencyMS,
		MaxFailedRequests: payload.MaxFailedRequests,
		Action:            strings.TrimSpace(payload.Action),
		ThrottleRatio:     payload.ThrottleRatio,
		Enabled:           payload.Enabled,
		Metadata:          payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) EvaluateSLOGuard(c *gin.Context) {
	if h == nil || h.sloGuardSvc == nil {
		contracts.ResponseServiceUnavailable(c, "slo guard service unavailable", nil)
		return
	}
	var payload evaluateSLOGuardRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sloGuardSvc.Evaluate(c.Request.Context(), tenantUUID, logisticssvc.SLOGuardEvaluateRequest{
		WindowHours: payload.WindowHours,
		CarrierID:   strings.TrimSpace(payload.CarrierID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) GetSLOGuardStatus(c *gin.Context) {
	if h == nil || h.sloGuardSvc == nil {
		contracts.ResponseServiceUnavailable(c, "slo guard service unavailable", nil)
		return
	}
	windowHours := 24
	if raw := strings.TrimSpace(c.Query("window_hours")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			windowHours = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sloGuardSvc.Status(c.Request.Context(), tenantUUID, logisticssvc.SLOGuardEvaluateRequest{
		WindowHours: windowHours,
		CarrierID:   strings.TrimSpace(c.Query("carrier_id")),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ReleaseSLOGuardPolicy(c *gin.Context) {
	if h == nil || h.sloGuardSvc == nil {
		contracts.ResponseServiceUnavailable(c, "slo guard service unavailable", nil)
		return
	}
	policyID := strings.TrimSpace(c.Param("id"))
	if policyID == "" {
		contracts.ResponseBadRequest(c, "policy id is required")
		return
	}
	var payload releaseSLOGuardRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sloGuardSvc.ManualRelease(c.Request.Context(), tenantUUID, logisticssvc.SLOGuardReleaseRequest{
		PolicyID:   policyID,
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
