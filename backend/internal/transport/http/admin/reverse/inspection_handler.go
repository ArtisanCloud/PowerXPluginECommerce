package reverse

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	reversesvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/reverse"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListInspectionRules(c *gin.Context) {
	if h == nil || h.inspectionSvc == nil {
		contracts.ResponseServiceUnavailable(c, "inspection service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.inspectionSvc.ListRules(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateInspectionRule(c *gin.Context) {
	if h == nil || h.inspectionSvc == nil {
		contracts.ResponseServiceUnavailable(c, "inspection service unavailable", nil)
		return
	}
	var payload createInspectionRuleRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.inspectionSvc.CreateRule(c.Request.Context(), tenantUUID, reversesvc.CreateInspectionRuleRequest{
		Name:           strings.TrimSpace(payload.Name),
		Priority:       payload.Priority,
		Condition:      payload.Condition,
		Decision:       strings.TrimSpace(payload.Decision),
		Recommendation: strings.TrimSpace(payload.Recommendation),
		Notes:          strings.TrimSpace(payload.Notes),
		Enabled:        payload.Enabled,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) EvaluateInspection(c *gin.Context) {
	if h == nil || h.inspectionSvc == nil {
		contracts.ResponseServiceUnavailable(c, "inspection service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	var payload evaluateInspectionRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.inspectionSvc.EvaluateWaybill(c.Request.Context(), tenantUUID, waybillID, reversesvc.EvaluateInspectionRequest{
		Attributes: payload.Attributes,
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Notes:      strings.TrimSpace(payload.Notes),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}
