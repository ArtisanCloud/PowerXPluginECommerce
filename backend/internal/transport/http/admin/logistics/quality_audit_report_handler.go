package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListQualityAuditReports(c *gin.Context) {
	if h == nil || h.qualityAuditSvc == nil {
		contracts.ResponseServiceUnavailable(c, "quality audit report service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.qualityAuditSvc.List(c.Request.Context(), tenantUUID, buildQualityAuditQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) GenerateQualityAuditReport(c *gin.Context) {
	if h == nil || h.qualityAuditSvc == nil {
		contracts.ResponseServiceUnavailable(c, "quality audit report service unavailable", nil)
		return
	}
	var payload generateQualityAuditReportRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.qualityAuditSvc.Generate(c.Request.Context(), tenantUUID, logisticssvc.GenerateQualityAuditReportRequest{
		CarrierID:        strings.TrimSpace(payload.CarrierID),
		WarehouseID:      strings.TrimSpace(payload.WarehouseID),
		DestinationZone:  strings.TrimSpace(payload.DestinationZone),
		ReportPeriodFrom: strings.TrimSpace(payload.ReportPeriodFrom),
		ReportPeriodTo:   strings.TrimSpace(payload.ReportPeriodTo),
		WindowHours:      payload.WindowHours,
		OperatorID:       strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) GetQualityAuditReport(c *gin.Context) {
	if h == nil || h.qualityAuditSvc == nil {
		contracts.ResponseServiceUnavailable(c, "quality audit report service unavailable", nil)
		return
	}
	reportID := strings.TrimSpace(c.Param("id"))
	if reportID == "" {
		contracts.ResponseBadRequest(c, "report id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.qualityAuditSvc.Get(c.Request.Context(), tenantUUID, reportID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ExportQualityAuditReport(c *gin.Context) {
	if h == nil || h.qualityAuditSvc == nil {
		contracts.ResponseServiceUnavailable(c, "quality audit report service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	content, err := h.qualityAuditSvc.ExportCSV(c.Request.Context(), tenantUUID, buildQualityAuditQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"content": content, "format": "csv"})
}

func buildQualityAuditQuery(c *gin.Context) logisticssvc.QualityAuditReportQuery {
	windowHours := 24
	if raw := strings.TrimSpace(c.Query("window_hours")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			windowHours = v
		}
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	return logisticssvc.QualityAuditReportQuery{
		CarrierID:       strings.TrimSpace(c.Query("carrier_id")),
		WarehouseID:     strings.TrimSpace(c.Query("warehouse_id")),
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		Status:          strings.TrimSpace(c.Query("status")),
		WindowHours:     windowHours,
		Limit:           limit,
	}
}
