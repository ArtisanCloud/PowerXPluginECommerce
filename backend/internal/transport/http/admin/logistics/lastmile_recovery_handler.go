package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListLastmileRecoveryRules(c *gin.Context) {
	if h == nil || h.lastmileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "lastmile recovery service unavailable", nil)
		return
	}
	var enabled *bool
	switch strings.ToLower(strings.TrimSpace(c.Query("enabled"))) {
	case "true":
		v := true
		enabled = &v
	case "false":
		v := false
		enabled = &v
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.lastmileSvc.ListRules(c.Request.Context(), tenantUUID, enabled)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertLastmileRecoveryRule(c *gin.Context) {
	if h == nil || h.lastmileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "lastmile recovery service unavailable", nil)
		return
	}
	var payload upsertLastmileRecoveryRuleRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.lastmileSvc.UpsertRule(c.Request.Context(), tenantUUID, logisticssvc.UpsertLastmileRecoveryRuleRequest{
		ID:           strings.TrimSpace(payload.ID),
		Name:         strings.TrimSpace(payload.Name),
		TriggerEvent: strings.TrimSpace(payload.TriggerEvent),
		Action:       strings.TrimSpace(payload.Action),
		Priority:     payload.Priority,
		MaxRetries:   payload.MaxRetries,
		Enabled:      payload.Enabled,
		Config:       payload.Config,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ExecuteLastmileRecovery(c *gin.Context) {
	if h == nil || h.lastmileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "lastmile recovery service unavailable", nil)
		return
	}
	var payload executeLastmileRecoveryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.lastmileSvc.Execute(c.Request.Context(), tenantUUID, logisticssvc.ExecuteLastmileRecoveryRequest{
		RequestKey:   strings.TrimSpace(payload.RequestKey),
		RuleID:       strings.TrimSpace(payload.RuleID),
		WaybillID:    strings.TrimSpace(payload.WaybillID),
		WaybillNo:    strings.TrimSpace(payload.WaybillNo),
		TriggerEvent: strings.TrimSpace(payload.TriggerEvent),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"run":                row,
		"idempotency_status": idem,
	})
}

func (h *Handler) ListLastmileRecoveryRuns(c *gin.Context) {
	if h == nil || h.lastmileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "lastmile recovery service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.lastmileSvc.ListRuns(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("waybill_no")),
		strings.TrimSpace(c.Query("status")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) TakeoverLastmileRecovery(c *gin.Context) {
	if h == nil || h.lastmileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "lastmile recovery service unavailable", nil)
		return
	}
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		contracts.ResponseBadRequest(c, "run id is required")
		return
	}
	var payload takeoverLastmileRecoveryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.lastmileSvc.Takeover(c.Request.Context(), tenantUUID, runID, logisticssvc.TakeoverLastmileRecoveryRequest{
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
