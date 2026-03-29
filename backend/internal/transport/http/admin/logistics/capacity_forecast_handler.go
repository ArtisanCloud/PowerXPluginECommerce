package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCapacityForecasts(c *gin.Context) {
	if h == nil || h.forecastSvc == nil {
		contracts.ResponseServiceUnavailable(c, "capacity forecast service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.forecastSvc.List(c.Request.Context(), tenantUUID, logisticssvc.CapacityForecastQuery{
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

func (h *Handler) GenerateCapacityForecast(c *gin.Context) {
	if h == nil || h.forecastSvc == nil {
		contracts.ResponseServiceUnavailable(c, "capacity forecast service unavailable", nil)
		return
	}
	var payload generateCapacityForecastRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.forecastSvc.Generate(c.Request.Context(), tenantUUID, logisticssvc.CapacityForecastGenerateRequest{
		CarrierID:       strings.TrimSpace(payload.CarrierID),
		WarehouseID:     strings.TrimSpace(payload.WarehouseID),
		DestinationZone: strings.TrimSpace(payload.DestinationZone),
		WindowDays:      payload.WindowDays,
		OperatorID:      strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ApplyCapacityForecast(c *gin.Context) {
	if h == nil || h.forecastSvc == nil {
		contracts.ResponseServiceUnavailable(c, "capacity forecast service unavailable", nil)
		return
	}
	forecastID := strings.TrimSpace(c.Param("id"))
	if forecastID == "" {
		contracts.ResponseBadRequest(c, "forecast id is required")
		return
	}
	var payload applyCapacityForecastRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.forecastSvc.Apply(c.Request.Context(), tenantUUID, logisticssvc.CapacityForecastApplyRequest{
		ForecastID: forecastID,
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
