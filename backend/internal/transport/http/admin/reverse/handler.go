package reverse

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	reversesvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/reverse"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	waybillSvc *reversesvc.WaybillService
}

func NewHandler(waybillSvc *reversesvc.WaybillService) *Handler {
	return &Handler{waybillSvc: waybillSvc}
}

func (h *Handler) ListWaybills(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reverse service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.waybillSvc.List(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateWaybill(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reverse service unavailable", nil)
		return
	}
	var payload createWaybillRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waybillSvc.Create(c.Request.Context(), tenantUUID, reversesvc.CreateWaybillRequest{
		OrderID:     strings.TrimSpace(payload.OrderID),
		AfterSaleID: strings.TrimSpace(payload.AfterSaleID),
		WaybillNo:   strings.TrimSpace(payload.WaybillNo),
		Metadata:    payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) GetWaybill(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reverse service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waybillSvc.Detail(c.Request.Context(), tenantUUID, waybillID)
	if err != nil {
		contracts.ResponseNotFound(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) AppendTracking(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reverse service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	var payload appendTrackingRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	var occurredAt *time.Time
	if payload.OccurredAt != nil {
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(*payload.OccurredAt))
		if err != nil {
			contracts.ResponseBadRequest(c, "occurred_at must be RFC3339")
			return
		}
		occurredAt = &ts
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waybillSvc.AppendTracking(c.Request.Context(), tenantUUID, waybillID, reversesvc.AppendTrackingRequest{
		Status:      strings.TrimSpace(payload.Status),
		Description: strings.TrimSpace(payload.Description),
		OccurredAt:  occurredAt,
		Payload:     payload.Payload,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RecordWarehouseResult(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reverse service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	var payload warehouseResultRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	record, wb, err := h.waybillSvc.RecordWarehouseResult(c.Request.Context(), tenantUUID, waybillID, reversesvc.RecordWarehouseResultRequest{
		Result:      strings.TrimSpace(payload.Result),
		Disposition: strings.TrimSpace(payload.Disposition),
		OperatorID:  strings.TrimSpace(payload.OperatorID),
		Notes:       strings.TrimSpace(payload.Notes),
		Metadata:    payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"warehouse_result": record, "waybill": wb})
}
