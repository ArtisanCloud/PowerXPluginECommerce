package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListSandboxScenarios(c *gin.Context) {
	if h == nil || h.sandboxSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment sandbox service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.sandboxSvc.ListScenarios(c.Request.Context(), tenantUUID, logisticssvc.FulfillmentSandboxScenarioQuery{
		CarrierID:       strings.TrimSpace(c.Query("carrier_id")),
		WarehouseID:     strings.TrimSpace(c.Query("warehouse_id")),
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		Status:          strings.TrimSpace(c.Query("status")),
		Limit:           limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertSandboxScenario(c *gin.Context) {
	if h == nil || h.sandboxSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment sandbox service unavailable", nil)
		return
	}
	var payload upsertFulfillmentSandboxScenarioRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sandboxSvc.UpsertScenario(c.Request.Context(), tenantUUID, logisticssvc.UpsertFulfillmentSandboxScenarioRequest{
		ID:              strings.TrimSpace(payload.ID),
		Name:            strings.TrimSpace(payload.Name),
		CarrierID:       strings.TrimSpace(payload.CarrierID),
		WarehouseID:     strings.TrimSpace(payload.WarehouseID),
		DestinationZone: strings.TrimSpace(payload.DestinationZone),
		BaselineConfig:  payload.BaselineConfig,
		StrategyConfig:  payload.StrategyConfig,
		Status:          strings.TrimSpace(payload.Status),
		Description:     strings.TrimSpace(payload.Description),
		OperatorID:      strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListSandboxRuns(c *gin.Context) {
	if h == nil || h.sandboxSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment sandbox service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.sandboxSvc.ListRuns(c.Request.Context(), tenantUUID, logisticssvc.FulfillmentSandboxRunQuery{
		ScenarioID: strings.TrimSpace(c.Query("scenario_id")),
		Strategy:   strings.TrimSpace(c.Query("strategy")),
		Limit:      limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) RunSandbox(c *gin.Context) {
	if h == nil || h.sandboxSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment sandbox service unavailable", nil)
		return
	}
	var payload runFulfillmentSandboxRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sandboxSvc.Run(c.Request.Context(), tenantUUID, logisticssvc.RunFulfillmentSandboxRequest{
		ScenarioID: strings.TrimSpace(payload.ScenarioID),
		WindowDays: payload.WindowDays,
		Strategy:   strings.TrimSpace(payload.Strategy),
		RequestKey: strings.TrimSpace(payload.RequestKey),
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) CompareSandboxRuns(c *gin.Context) {
	if h == nil || h.sandboxSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment sandbox service unavailable", nil)
		return
	}
	var payload compareFulfillmentSandboxRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.sandboxSvc.Compare(c.Request.Context(), tenantUUID, logisticssvc.CompareFulfillmentSandboxRequest{
		BaselineRunID:  strings.TrimSpace(payload.BaselineRunID),
		CandidateRunID: strings.TrimSpace(payload.CandidateRunID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
