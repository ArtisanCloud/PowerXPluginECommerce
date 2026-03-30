package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetTrackingRootCauseSummary(c *gin.Context) {
	if h == nil || h.rootCauseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "tracking root cause service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	summary, err := h.rootCauseSvc.Summary(c.Request.Context(), tenantUUID, buildTrackingRootCauseQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, summary)
}

func (h *Handler) ListTrackingRootCauses(c *gin.Context) {
	if h == nil || h.rootCauseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "tracking root cause service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.rootCauseSvc.List(c.Request.Context(), tenantUUID, buildTrackingRootCauseQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) AnalyzeTrackingRootCauses(c *gin.Context) {
	if h == nil || h.rootCauseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "tracking root cause service unavailable", nil)
		return
	}
	var payload analyzeTrackingRootCauseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.rootCauseSvc.Analyze(c.Request.Context(), tenantUUID, logisticssvc.AnalyzeTrackingRootCauseRequest{
		CarrierID:       strings.TrimSpace(payload.CarrierID),
		WarehouseID:     strings.TrimSpace(payload.WarehouseID),
		DestinationZone: strings.TrimSpace(payload.DestinationZone),
		WindowHours:     payload.WindowHours,
		Limit:           payload.Limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) HandleTrackingRootCause(c *gin.Context) {
	if h == nil || h.rootCauseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "tracking root cause service unavailable", nil)
		return
	}
	rootCauseID := strings.TrimSpace(c.Param("id"))
	if rootCauseID == "" {
		contracts.ResponseBadRequest(c, "root cause id is required")
		return
	}
	var payload handleTrackingRootCauseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.rootCauseSvc.Handle(c.Request.Context(), tenantUUID, logisticssvc.HandleTrackingRootCauseRequest{
		RootCauseID: rootCauseID,
		Action:      strings.TrimSpace(payload.Action),
		Status:      strings.TrimSpace(payload.Status),
		OperatorID:  strings.TrimSpace(payload.OperatorID),
		ResultNote:  strings.TrimSpace(payload.ResultNote),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func buildTrackingRootCauseQuery(c *gin.Context) logisticssvc.TrackingRootCauseQuery {
	windowHours := 24
	if raw := strings.TrimSpace(c.Query("window_hours")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			windowHours = v
		}
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	return logisticssvc.TrackingRootCauseQuery{
		CarrierID:       strings.TrimSpace(c.Query("carrier_id")),
		WarehouseID:     strings.TrimSpace(c.Query("warehouse_id")),
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		AnomalyType:     strings.TrimSpace(c.Query("anomaly_type")),
		Status:          strings.TrimSpace(c.Query("status")),
		WindowHours:     windowHours,
		Limit:           limit,
	}
}
