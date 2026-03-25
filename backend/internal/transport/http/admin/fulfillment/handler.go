package fulfillment

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	fulfillmentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/fulfillment"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskSvc      *fulfillmentsvc.TaskService
	exceptionSvc *fulfillmentsvc.ExceptionService
}

func NewHandler(taskSvc *fulfillmentsvc.TaskService, exceptionSvc *fulfillmentsvc.ExceptionService) *Handler {
	return &Handler{taskSvc: taskSvc, exceptionSvc: exceptionSvc}
}

func (h *Handler) ListTasks(c *gin.Context) {
	if h == nil || h.taskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment task service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.taskSvc.List(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("status")))
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateTask(c *gin.Context) {
	if h == nil || h.taskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment task service unavailable", nil)
		return
	}
	var payload createTaskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.taskSvc.Create(c.Request.Context(), tenantUUID, fulfillmentsvc.CreateTaskRequest{
		OrderID:     strings.TrimSpace(payload.OrderID),
		WarehouseID: strings.TrimSpace(payload.WarehouseID),
		AssignedTo:  strings.TrimSpace(payload.AssignedTo),
		Metadata:    payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) CompleteTask(c *gin.Context) {
	if h == nil || h.taskSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment task service unavailable", nil)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	var payload completeTaskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.taskSvc.Complete(c.Request.Context(), tenantUUID, taskID, fulfillmentsvc.AdvanceTaskRequest{
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Detail:     payload.Detail,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListExceptions(c *gin.Context) {
	if h == nil || h.exceptionSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment exception service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	status := strings.TrimSpace(c.Query("status"))
	rows, err := h.exceptionSvc.List(c.Request.Context(), tenantUUID, status)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ReportException(c *gin.Context) {
	if h == nil || h.exceptionSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment exception service unavailable", nil)
		return
	}
	var payload reportExceptionRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.exceptionSvc.Report(c.Request.Context(), tenantUUID, fulfillmentsvc.ReportExceptionRequest{
		TaskID:    strings.TrimSpace(payload.TaskID),
		WaybillID: strings.TrimSpace(payload.WaybillID),
		Type:      strings.TrimSpace(payload.Type),
		Reason:    strings.TrimSpace(payload.Reason),
		Metadata:  payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	_, _ = h.exceptionSvc.EscalateOverdue(c.Request.Context(), tenantUUID, time.Now().UTC())
	contracts.ResponseSuccess(c, row)
}
