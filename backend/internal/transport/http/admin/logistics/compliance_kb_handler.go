package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListComplianceKBVersions(c *gin.Context) {
	if h == nil || h.complianceKBSvc == nil {
		contracts.ResponseServiceUnavailable(c, "compliance kb service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.complianceKBSvc.ListPolicies(c.Request.Context(), tenantUUID, logisticssvc.ComplianceKBQuery{
		CountryCode: strings.TrimSpace(c.Query("country_code")),
		Status:      strings.TrimSpace(c.Query("status")),
		Limit:       limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) SyncComplianceKBPolicy(c *gin.Context) {
	if h == nil || h.complianceKBSvc == nil {
		contracts.ResponseServiceUnavailable(c, "compliance kb service unavailable", nil)
		return
	}
	var payload syncComplianceKBPolicyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.complianceKBSvc.SyncPolicy(c.Request.Context(), tenantUUID, logisticssvc.SyncComplianceKBPolicyRequest{
		CountryCode:     strings.TrimSpace(payload.CountryCode),
		PackID:          strings.TrimSpace(payload.PackID),
		SourceVersionID: strings.TrimSpace(payload.SourceVersionID),
		PolicyVersion:   strings.TrimSpace(payload.PolicyVersion),
		EffectiveFrom:   strings.TrimSpace(payload.EffectiveFrom),
		EffectiveTo:     strings.TrimSpace(payload.EffectiveTo),
		Notes:           strings.TrimSpace(payload.Notes),
		OperatorID:      strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) DiffComplianceKBPolicies(c *gin.Context) {
	if h == nil || h.complianceKBSvc == nil {
		contracts.ResponseServiceUnavailable(c, "compliance kb service unavailable", nil)
		return
	}
	var payload diffComplianceKBPoliciesRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.complianceKBSvc.DiffPolicies(c.Request.Context(), tenantUUID, logisticssvc.DiffComplianceKBPolicyRequest{
		BasePolicyID:   strings.TrimSpace(payload.BasePolicyID),
		TargetPolicyID: strings.TrimSpace(payload.TargetPolicyID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) PublishComplianceKBPolicy(c *gin.Context) {
	if h == nil || h.complianceKBSvc == nil {
		contracts.ResponseServiceUnavailable(c, "compliance kb service unavailable", nil)
		return
	}
	policyID := strings.TrimSpace(c.Param("id"))
	if policyID == "" {
		contracts.ResponseBadRequest(c, "policy id is required")
		return
	}
	var payload publishComplianceKBPolicyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.complianceKBSvc.PublishPolicy(c.Request.Context(), tenantUUID, policyID, logisticssvc.PublishComplianceKBPolicyRequest{
		RolloutPercent: payload.RolloutPercent,
		RolloutTenants: payload.RolloutTenants,
		EffectiveFrom:  strings.TrimSpace(payload.EffectiveFrom),
		EffectiveTo:    strings.TrimSpace(payload.EffectiveTo),
		OperatorID:     strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
