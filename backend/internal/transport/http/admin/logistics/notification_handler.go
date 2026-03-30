package logistics

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListNotificationTemplates(c *gin.Context) {
	if h == nil || h.notifySvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics notification service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.notifySvc.ListTemplates(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertNotificationTemplate(c *gin.Context) {
	if h == nil || h.notifySvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics notification service unavailable", nil)
		return
	}
	var payload upsertNotificationTemplateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.notifySvc.UpsertTemplate(c.Request.Context(), tenantUUID, logisticssvc.UpsertNotificationTemplateRequest{
		ID:       strings.TrimSpace(payload.ID),
		Name:     strings.TrimSpace(payload.Name),
		Event:    strings.TrimSpace(payload.Event),
		Channel:  strings.TrimSpace(payload.Channel),
		Title:    strings.TrimSpace(payload.Title),
		Body:     strings.TrimSpace(payload.Body),
		Enabled:  payload.Enabled,
		Metadata: payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListNotificationRecords(c *gin.Context) {
	if h == nil || h.notifySvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics notification service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.notifySvc.ListRecords(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("status")))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) SendNotification(c *gin.Context) {
	if h == nil || h.notifySvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics notification service unavailable", nil)
		return
	}
	var payload sendNotificationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.notifySvc.Send(c.Request.Context(), tenantUUID, logisticssvc.SendNotificationRequest{
		WaybillID:      strings.TrimSpace(payload.WaybillID),
		Event:          strings.TrimSpace(payload.Event),
		IdempotencyKey: strings.TrimSpace(payload.IdempotencyKey),
		Payload:        payload.Payload,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RetryNotification(c *gin.Context) {
	if h == nil || h.notifySvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics notification service unavailable", nil)
		return
	}
	recordID := strings.TrimSpace(c.Param("id"))
	var payload retryNotificationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if recordID == "" {
		recordID = strings.TrimSpace(payload.RecordID)
	}
	if recordID == "" {
		contracts.ResponseBadRequest(c, "record id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.notifySvc.Retry(c.Request.Context(), tenantUUID, logisticssvc.RetryNotificationRequest{
		RecordID: recordID,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
