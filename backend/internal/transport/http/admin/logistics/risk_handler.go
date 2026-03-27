package logistics

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListRiskRules(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.riskSvc.ListRules(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertRiskRule(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	var payload upsertRiskRuleRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.riskSvc.UpsertRule(c.Request.Context(), tenantUUID, logisticssvc.UpsertRiskRuleRequest{
		ID:          strings.TrimSpace(payload.ID),
		Name:        strings.TrimSpace(payload.Name),
		MatchField:  strings.TrimSpace(payload.MatchField),
		MatchMode:   strings.TrimSpace(payload.MatchMode),
		Pattern:     strings.TrimSpace(payload.Pattern),
		Decision:    strings.TrimSpace(payload.Decision),
		RiskLevel:   strings.TrimSpace(payload.RiskLevel),
		Priority:    payload.Priority,
		Enabled:     payload.Enabled,
		Description: strings.TrimSpace(payload.Description),
		RuleConfig:  payload.RuleConfig,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListRiskBlacklist(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.riskSvc.ListBlacklist(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("status")))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertRiskBlacklist(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	var payload upsertBlacklistEntryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	var expiresAt *time.Time
	if payload.ExpiresAt != nil && strings.TrimSpace(*payload.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*payload.ExpiresAt))
		if err != nil {
			contracts.ResponseBadRequest(c, "expires_at must be RFC3339")
			return
		}
		expiresAt = &parsed
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.riskSvc.UpsertBlacklist(c.Request.Context(), tenantUUID, logisticssvc.UpsertBlacklistEntryRequest{
		ID:             strings.TrimSpace(payload.ID),
		EntryType:      strings.TrimSpace(payload.EntryType),
		RecipientName:  strings.TrimSpace(payload.RecipientName),
		RecipientPhone: strings.TrimSpace(payload.RecipientPhone),
		AddressLine:    strings.TrimSpace(payload.AddressLine),
		Reason:         strings.TrimSpace(payload.Reason),
		Status:         strings.TrimSpace(payload.Status),
		ExpiresAt:      expiresAt,
		Metadata:       payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListRiskHits(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.riskSvc.ListHits(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("waybill_id")),
		strings.TrimSpace(c.Query("status")),
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) EvaluateRisk(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	var payload evaluateRiskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.riskSvc.Evaluate(c.Request.Context(), tenantUUID, logisticssvc.EvaluateRiskRequest{
		WaybillID:       strings.TrimSpace(payload.WaybillID),
		WaybillNo:       strings.TrimSpace(payload.WaybillNo),
		RecipientName:   strings.TrimSpace(payload.RecipientName),
		RecipientPhone:  strings.TrimSpace(payload.RecipientPhone),
		DestinationLine: strings.TrimSpace(payload.DestinationLine),
		OperatorID:      strings.TrimSpace(payload.OperatorID),
		Context:         payload.Context,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}

func (h *Handler) ReleaseRiskHit(c *gin.Context) {
	if h == nil || h.riskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics risk service unavailable", nil)
		return
	}
	hitID := strings.TrimSpace(c.Param("id"))
	if hitID == "" {
		contracts.ResponseBadRequest(c, "hit id is required")
		return
	}
	var payload releaseRiskHitRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.riskSvc.Release(c.Request.Context(), tenantUUID, hitID, logisticssvc.ReleaseRiskHitRequest{
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
