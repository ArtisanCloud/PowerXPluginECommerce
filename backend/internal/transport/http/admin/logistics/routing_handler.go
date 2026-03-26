package logistics

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListRoutingRules(c *gin.Context) {
	if h == nil || h.routingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics routing service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.routingSvc.ListRules(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertRoutingRule(c *gin.Context) {
	if h == nil || h.routingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics routing service unavailable", nil)
		return
	}
	var payload upsertRoutingRuleRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.routingSvc.UpsertRule(c.Request.Context(), tenantUUID, logisticssvc.UpsertRoutingRuleRequest{
		ID:              strings.TrimSpace(payload.ID),
		Name:            strings.TrimSpace(payload.Name),
		WarehouseID:     strings.TrimSpace(payload.WarehouseID),
		DestinationZone: strings.TrimSpace(payload.DestinationZone),
		CarrierID:       strings.TrimSpace(payload.CarrierID),
		ServiceCode:     strings.TrimSpace(payload.ServiceCode),
		Priority:        payload.Priority,
		MinWeight:       payload.MinWeight,
		MaxWeight:       payload.MaxWeight,
		MinOrderAmount:  payload.MinOrderAmount,
		MaxOrderAmount:  payload.MaxOrderAmount,
		Fallback:        payload.Fallback,
		Enabled:         payload.Enabled,
		RuleConfig:      payload.RuleConfig,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) DeleteRoutingRule(c *gin.Context) {
	if h == nil || h.routingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics routing service unavailable", nil)
		return
	}
	ruleID := strings.TrimSpace(c.Param("id"))
	if ruleID == "" {
		contracts.ResponseBadRequest(c, "rule id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if err := h.routingSvc.DeleteRule(c.Request.Context(), tenantUUID, ruleID); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"id": ruleID, "deleted": true})
}

func (h *Handler) PreviewRouting(c *gin.Context) {
	if h == nil || h.routingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics routing service unavailable", nil)
		return
	}
	var payload previewRoutingRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.routingSvc.Preview(c.Request.Context(), tenantUUID, logisticssvc.PreviewRoutingRequest{
		WarehouseID:        strings.TrimSpace(payload.WarehouseID),
		DestinationZone:    strings.TrimSpace(payload.DestinationZone),
		Weight:             payload.Weight,
		OrderAmount:        payload.OrderAmount,
		PreferredCarrierID: strings.TrimSpace(payload.PreferredCarrierID),
		ServiceCode:        strings.TrimSpace(payload.ServiceCode),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}
