package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetKPIDashboardOverview(c *gin.Context) {
	if h == nil || h.kpiSvc == nil {
		contracts.ResponseServiceUnavailable(c, "kpi dashboard service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.kpiSvc.Overview(c.Request.Context(), tenantUUID, buildKPIDashboardQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) GetKPIDashboardTrends(c *gin.Context) {
	if h == nil || h.kpiSvc == nil {
		contracts.ResponseServiceUnavailable(c, "kpi dashboard service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.kpiSvc.Trends(c.Request.Context(), tenantUUID, buildKPIDashboardQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) GetKPIDashboardDrilldown(c *gin.Context) {
	if h == nil || h.kpiSvc == nil {
		contracts.ResponseServiceUnavailable(c, "kpi dashboard service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.kpiSvc.Drilldown(c.Request.Context(), tenantUUID, buildKPIDashboardQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ExportKPIDashboard(c *gin.Context) {
	if h == nil || h.kpiSvc == nil {
		contracts.ResponseServiceUnavailable(c, "kpi dashboard service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	content, err := h.kpiSvc.ExportCSV(c.Request.Context(), tenantUUID, buildKPIDashboardQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"content": content, "format": "csv"})
}

func buildKPIDashboardQuery(c *gin.Context) logisticssvc.KPIDashboardQuery {
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
	return logisticssvc.KPIDashboardQuery{
		WindowHours:     windowHours,
		Dimension:       strings.TrimSpace(c.Query("dimension")),
		CarrierID:       strings.TrimSpace(c.Query("carrier_id")),
		WarehouseID:     strings.TrimSpace(c.Query("warehouse_id")),
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		Limit:           limit,
	}
}
