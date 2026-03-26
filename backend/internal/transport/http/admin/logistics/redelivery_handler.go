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

func (h *Handler) ListRedeliveryTasks(c *gin.Context) {
	if h == nil || h.redelivSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics redelivery service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.redelivSvc.List(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("waybill_id")),
		strings.TrimSpace(c.Query("status")),
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) InitiateRedeliveryTask(c *gin.Context) {
	if h == nil || h.redelivSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics redelivery service unavailable", nil)
		return
	}
	var payload initiateRedeliveryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.redelivSvc.Initiate(c.Request.Context(), tenantUUID, logisticssvc.InitiateRedeliveryRequest{
		WaybillID:  strings.TrimSpace(payload.WaybillID),
		RequestKey: strings.TrimSpace(payload.RequestKey),
		Reason:     strings.TrimSpace(payload.Reason),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Address:    payload.Address,
		Metadata:   payload.Metadata,
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
		"data":       gin.H{"task": row, "idempotency_status": idem},
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

func (h *Handler) UpdateRedeliveryAddress(c *gin.Context) {
	if h == nil || h.redelivSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics redelivery service unavailable", nil)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	var payload updateRedeliveryAddressRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.redelivSvc.UpdateAddress(c.Request.Context(), tenantUUID, taskID, logisticssvc.UpdateRedeliveryAddressRequest{
		Address:    payload.Address,
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RedispatchRedeliveryTask(c *gin.Context) {
	if h == nil || h.redelivSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics redelivery service unavailable", nil)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	var payload redispatchRedeliveryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.redelivSvc.Redispatch(c.Request.Context(), tenantUUID, taskID, logisticssvc.RedispatchRedeliveryRequest{
		RequestKey: strings.TrimSpace(payload.RequestKey),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
		Metadata:   payload.Metadata,
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
		"data":       gin.H{"task": row, "idempotency_status": idem},
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
	})
}

func (h *Handler) CloseRedeliveryTask(c *gin.Context) {
	if h == nil || h.redelivSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics redelivery service unavailable", nil)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	var payload closeRedeliveryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.redelivSvc.Close(c.Request.Context(), tenantUUID, taskID, logisticssvc.CloseRedeliveryRequest{
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Reason:     strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
