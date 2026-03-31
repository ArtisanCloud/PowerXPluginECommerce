package subscription_reconciliation

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	SubscriptionReconciliationSvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/subscription_reconciliation"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RunGovernance(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	var payload runGovernanceRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	payload = payload.normalize()
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.RunGovernance(c.Request.Context(), tenantUUID, SubscriptionReconciliationSvc.RunGovernanceInput{
		BillingCycle: payload.BillingCycle,
		DryRun:       payload.DryRun,
	})
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}
