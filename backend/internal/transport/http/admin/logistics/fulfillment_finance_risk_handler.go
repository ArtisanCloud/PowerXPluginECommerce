package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListFinanceRisks(c *gin.Context) {
	if h == nil || h.financeRiskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "finance risk service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.financeRiskSvc.ListRisks(c.Request.Context(), tenantUUID, logisticssvc.FinanceRiskQuery{
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		Status:    strings.TrimSpace(c.Query("status")),
		RiskLevel: strings.TrimSpace(c.Query("risk_level")),
		Limit:     limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) EvaluateFinanceRisks(c *gin.Context) {
	if h == nil || h.financeRiskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "finance risk service unavailable", nil)
		return
	}
	var payload evaluateFinanceRiskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.financeRiskSvc.Evaluate(c.Request.Context(), tenantUUID, logisticssvc.EvaluateFinanceRiskRequest{
		CarrierID: strings.TrimSpace(payload.CarrierID),
		Threshold: payload.Threshold,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ExecuteFinanceRiskAction(c *gin.Context) {
	if h == nil || h.financeRiskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "finance risk service unavailable", nil)
		return
	}
	riskID := strings.TrimSpace(c.Param("id"))
	if riskID == "" {
		contracts.ResponseBadRequest(c, "risk id is required")
		return
	}
	var payload executeFinanceRiskActionRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.financeRiskSvc.ExecuteAction(c.Request.Context(), tenantUUID, riskID, logisticssvc.ExecuteFinanceRiskActionRequest{
		Action:     strings.TrimSpace(payload.Action),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Note:       strings.TrimSpace(payload.Note),
		RequestKey: strings.TrimSpace(payload.RequestKey),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"risk": row, "idempotency_status": idem})
}

func (h *Handler) ListFinanceRiskAudits(c *gin.Context) {
	if h == nil || h.financeRiskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "finance risk service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.financeRiskSvc.ListAudits(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("risk_id")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}
