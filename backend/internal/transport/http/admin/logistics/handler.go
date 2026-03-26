package logistics

import (
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	carrierSvc *logisticssvc.CarrierService
	rateSvc    *logisticssvc.RateTemplateService
	waybillSvc *logisticssvc.WaybillService
	billingSvc *logisticssvc.BillingService
	webhookSvc *logisticssvc.WebhookService
}

func NewHandler(
	carrierSvc *logisticssvc.CarrierService,
	rateSvc *logisticssvc.RateTemplateService,
	waybillSvc *logisticssvc.WaybillService,
	billingSvc *logisticssvc.BillingService,
	webhookSvc *logisticssvc.WebhookService,
) *Handler {
	return &Handler{carrierSvc: carrierSvc, rateSvc: rateSvc, waybillSvc: waybillSvc, billingSvc: billingSvc, webhookSvc: webhookSvc}
}

func (h *Handler) ListCarriers(c *gin.Context) {
	if h == nil || h.carrierSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics carrier service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.carrierSvc.List(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertCarrier(c *gin.Context) {
	if h == nil || h.carrierSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics carrier service unavailable", nil)
		return
	}
	var payload upsertCarrierRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.carrierSvc.Upsert(c.Request.Context(), tenantUUID, logisticssvc.UpsertCarrierRequest{
		ID:           strings.TrimSpace(payload.ID),
		Name:         strings.TrimSpace(payload.Name),
		Code:         strings.TrimSpace(payload.Code),
		Type:         strings.TrimSpace(payload.Type),
		Status:       strings.TrimSpace(payload.Status),
		ContactName:  strings.TrimSpace(payload.ContactName),
		ContactPhone: strings.TrimSpace(payload.ContactPhone),
		Capabilities: payload.Capabilities,
		Config:       payload.Config,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) DisableCarrier(c *gin.Context) {
	if h == nil || h.carrierSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics carrier service unavailable", nil)
		return
	}
	carrierID := strings.TrimSpace(c.Param("id"))
	if carrierID == "" {
		contracts.ResponseBadRequest(c, "carrier id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.carrierSvc.Disable(c.Request.Context(), tenantUUID, carrierID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) TestCarrier(c *gin.Context) {
	if h == nil || h.carrierSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics carrier service unavailable", nil)
		return
	}
	carrierID := strings.TrimSpace(c.Param("id"))
	if carrierID == "" {
		contracts.ResponseBadRequest(c, "carrier id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.carrierSvc.TestConnectivity(c.Request.Context(), tenantUUID, carrierID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	if h == nil || h.rateSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics template service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.rateSvc.List(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertTemplate(c *gin.Context) {
	if h == nil || h.rateSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics template service unavailable", nil)
		return
	}
	var payload upsertRateTemplateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.rateSvc.Upsert(c.Request.Context(), tenantUUID, logisticssvc.UpsertRateTemplateRequest{
		ID:       strings.TrimSpace(payload.ID),
		Name:     strings.TrimSpace(payload.Name),
		Currency: strings.TrimSpace(payload.Currency),
		Channels: payload.Channels,
		Rules:    payload.Rules,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) PublishTemplate(c *gin.Context) {
	if h == nil || h.rateSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics template service unavailable", nil)
		return
	}
	templateID := strings.TrimSpace(c.Param("id"))
	if templateID == "" {
		contracts.ResponseBadRequest(c, "template id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.rateSvc.Publish(c.Request.Context(), tenantUUID, templateID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListWaybills(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics waybill service unavailable", nil)
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
		contracts.ResponseServiceUnavailable(c, "logistics waybill service unavailable", nil)
		return
	}
	var payload createWaybillRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	waybill, idem, err := h.waybillSvc.Create(c.Request.Context(), tenantUUID, logisticssvc.CreateWaybillRequest{
		OrderID:              strings.TrimSpace(payload.OrderID),
		CarrierID:            strings.TrimSpace(payload.CarrierID),
		ServiceCode:          strings.TrimSpace(payload.ServiceCode),
		WaybillNo:            strings.TrimSpace(payload.WaybillNo),
		ManualFallbackReason: strings.TrimSpace(payload.ManualFallbackReason),
		PackageNo:            payload.PackageNo,
		PackageKey:           strings.TrimSpace(payload.PackageKey),
		ShipmentItems:        payload.ShipmentItems,
		OrderItemCount:       payload.OrderItemCount,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseCreated(c, gin.H{"waybill": waybill, "idempotency_status": idem})
}

func (h *Handler) GetWaybill(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics waybill service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	detail, err := h.waybillSvc.Detail(c.Request.Context(), tenantUUID, waybillID)
	if err != nil {
		contracts.ResponseNotFound(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func (h *Handler) AppendTracking(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics waybill service unavailable", nil)
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
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*payload.OccurredAt))
		if err != nil {
			contracts.ResponseBadRequest(c, "occurred_at must be RFC3339")
			return
		}
		occurredAt = &t
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	event, idem, err := h.waybillSvc.AppendTracking(c.Request.Context(), tenantUUID, waybillID, logisticssvc.AppendTrackingRequest{
		EventID:     strings.TrimSpace(payload.EventID),
		Status:      strings.TrimSpace(payload.Status),
		Source:      strings.TrimSpace(payload.Source),
		Description: strings.TrimSpace(payload.Description),
		OccurredAt:  occurredAt,
		Payload:     payload.Payload,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	statusCode := http.StatusCreated
	if idem == "replayed" {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, gin.H{
		"success":    true,
		"data":       gin.H{"tracking": event, "idempotency_status": idem},
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

func (h *Handler) CancelWaybill(c *gin.Context) {
	if h == nil || h.waybillSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics waybill service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waybillSvc.Cancel(c.Request.Context(), tenantUUID, waybillID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) UpdateWaybillCost(c *gin.Context) {
	if h == nil || h.billingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill id is required")
		return
	}
	var payload updateWaybillCostRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.billingSvc.UpdateWaybillCost(c.Request.Context(), tenantUUID, waybillID, logisticssvc.UpdateWaybillCostRequest{
		ActualFeeAmount: payload.ActualFeeAmount,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) BillingSummary(c *gin.Context) {
	if h == nil || h.billingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	query := logisticssvc.BillingQuery{
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		From:      strings.TrimSpace(c.Query("from")),
		To:        strings.TrimSpace(c.Query("to")),
	}
	snapshot, err := h.billingSvc.Snapshot(c.Request.Context(), tenantUUID, query)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, snapshot)
}

func (h *Handler) ExportBilling(c *gin.Context) {
	if h == nil || h.billingSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics billing service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	query := logisticssvc.BillingQuery{
		CarrierID: strings.TrimSpace(c.Query("carrier_id")),
		From:      strings.TrimSpace(c.Query("from")),
		To:        strings.TrimSpace(c.Query("to")),
	}
	payload, err := h.billingSvc.ExportPayload(c.Request.Context(), tenantUUID, query, strings.TrimSpace(c.Query("format")))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, payload)
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	if h == nil || h.webhookSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics webhook service unavailable", nil)
		return
	}
	var payload webhookRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	var occurredAt *time.Time
	if payload.OccurredAt != nil {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*payload.OccurredAt))
		if err != nil {
			contracts.ResponseBadRequest(c, "occurred_at must be RFC3339")
			return
		}
		occurredAt = &t
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	ack, err := h.webhookSvc.Handle(c.Request.Context(), tenantUUID, logisticssvc.HandleWebhookRequest{
		EventID:    strings.TrimSpace(payload.EventID),
		WaybillNo:  strings.TrimSpace(payload.WaybillNo),
		Status:     strings.TrimSpace(payload.Status),
		OccurredAt: occurredAt,
		Payload:    payload.Payload,
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadGateway, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, ack)
}
