package logistics

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListBillingCases(c *gin.Context) {
	if h == nil || h.caseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing case service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.caseSvc.List(c.Request.Context(), tenantUUID, logisticssvc.BillingCaseQuery{
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		Status:    strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateBillingCase(c *gin.Context) {
	if h == nil || h.caseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing case service unavailable", nil)
		return
	}
	var payload createBillingCaseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.caseSvc.Create(c.Request.Context(), tenantUUID, logisticssvc.CreateBillingCaseRequest{
		WaybillID: strings.TrimSpace(payload.WaybillID),
		Reason:    strings.TrimSpace(payload.Reason),
		Metadata:  payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"case": row, "idempotency_status": idem})
}

func (h *Handler) TransitionBillingCase(c *gin.Context) {
	if h == nil || h.caseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing case service unavailable", nil)
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	if caseID == "" {
		contracts.ResponseBadRequest(c, "case id is required")
		return
	}
	var payload transitionBillingCaseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.caseSvc.Transition(c.Request.Context(), tenantUUID, caseID, logisticssvc.TransitionBillingCaseRequest{
		Action:     strings.TrimSpace(payload.Action),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Note:       strings.TrimSpace(payload.Note),
		Metadata:   payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
