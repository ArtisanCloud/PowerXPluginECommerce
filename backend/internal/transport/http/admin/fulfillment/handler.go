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
	waveSvc      *fulfillmentsvc.WaveService
	strategySvc  *fulfillmentsvc.WaveStrategyService
	exceptionSvc *fulfillmentsvc.ExceptionService
	warehouseSvc *fulfillmentsvc.WarehouseBridgeService
}

func NewHandler(
	taskSvc *fulfillmentsvc.TaskService,
	waveSvc *fulfillmentsvc.WaveService,
	strategySvc *fulfillmentsvc.WaveStrategyService,
	exceptionSvc *fulfillmentsvc.ExceptionService,
	warehouseSvc *fulfillmentsvc.WarehouseBridgeService,
) *Handler {
	return &Handler{
		taskSvc:      taskSvc,
		waveSvc:      waveSvc,
		strategySvc:  strategySvc,
		exceptionSvc: exceptionSvc,
		warehouseSvc: warehouseSvc,
	}
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

func (h *Handler) ListWaves(c *gin.Context) {
	if h == nil || h.waveSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.waveSvc.List(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("status")))
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateWave(c *gin.Context) {
	if h == nil || h.waveSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave service unavailable", nil)
		return
	}
	var payload createWaveRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waveSvc.Create(c.Request.Context(), tenantUUID, fulfillmentsvc.CreateWaveRequest{
		Name:        strings.TrimSpace(payload.Name),
		WarehouseID: strings.TrimSpace(payload.WarehouseID),
		TaskIDs:     payload.TaskIDs,
		Metadata:    payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) GetWave(c *gin.Context) {
	if h == nil || h.waveSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave service unavailable", nil)
		return
	}
	waveID := strings.TrimSpace(c.Param("id"))
	if waveID == "" {
		contracts.ResponseBadRequest(c, "wave id is required")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waveSvc.Detail(c.Request.Context(), tenantUUID, waveID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) AdvanceWave(c *gin.Context) {
	if h == nil || h.waveSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave service unavailable", nil)
		return
	}
	waveID := strings.TrimSpace(c.Param("id"))
	if waveID == "" {
		contracts.ResponseBadRequest(c, "wave id is required")
		return
	}
	var payload advanceWaveRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waveSvc.BatchAdvance(c.Request.Context(), tenantUUID, waveID, fulfillmentsvc.BatchAdvanceWaveRequest{
		Status:      strings.TrimSpace(payload.Status),
		OperatorID:  strings.TrimSpace(payload.OperatorID),
		FailTaskIDs: payload.FailTaskIDs,
		Reason:      strings.TrimSpace(payload.Reason),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ReassignWaveTask(c *gin.Context) {
	if h == nil || h.waveSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave service unavailable", nil)
		return
	}
	waveID := strings.TrimSpace(c.Param("id"))
	taskID := strings.TrimSpace(c.Param("task_id"))
	if waveID == "" || taskID == "" {
		contracts.ResponseBadRequest(c, "wave id/task id is required")
		return
	}
	var payload reassignWaveTaskRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.waveSvc.ReassignTask(c.Request.Context(), tenantUUID, waveID, fulfillmentsvc.ReassignWaveTaskRequest{
		TaskID:     taskID,
		AssignedTo: strings.TrimSpace(payload.AssignedTo),
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListWaveStrategies(c *gin.Context) {
	if h == nil || h.strategySvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave strategy service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.strategySvc.List(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateWaveStrategy(c *gin.Context) {
	if h == nil || h.strategySvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave strategy service unavailable", nil)
		return
	}
	var payload createWaveStrategyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.strategySvc.Create(c.Request.Context(), tenantUUID, fulfillmentsvc.CreateWaveStrategyRequest{
		Name:            strings.TrimSpace(payload.Name),
		WarehouseID:     strings.TrimSpace(payload.WarehouseID),
		CarrierCode:     strings.TrimSpace(payload.CarrierCode),
		TimeWindow:      strings.TrimSpace(payload.TimeWindow),
		PriorityBand:    strings.TrimSpace(payload.PriorityBand),
		MaxTasksPerWave: payload.MaxTasksPerWave,
		Enabled:         payload.Enabled,
		Rules:           payload.Rules,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) PreviewWaveStrategy(c *gin.Context) {
	if h == nil || h.strategySvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment wave strategy service unavailable", nil)
		return
	}
	var payload previewWaveStrategyRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.strategySvc.Preview(c.Request.Context(), tenantUUID, fulfillmentsvc.PreviewWaveStrategyRequest{
		StrategyID: strings.TrimSpace(payload.StrategyID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListOutbounds(c *gin.Context) {
	if h == nil || h.warehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment warehouse service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.warehouseSvc.ListOutbounds(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("status")))
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateOutbound(c *gin.Context) {
	if h == nil || h.warehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment warehouse service unavailable", nil)
		return
	}
	var payload createOutboundRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	items := make([]fulfillmentsvc.PickLineItem, 0, len(payload.Items))
	for _, item := range payload.Items {
		items = append(items, fulfillmentsvc.PickLineItem{
			SKU: strings.TrimSpace(item.SKU),
			Qty: item.Qty,
		})
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.warehouseSvc.CreateOutbound(c.Request.Context(), tenantUUID, fulfillmentsvc.CreateOutboundRequest{
		TaskID:    strings.TrimSpace(payload.TaskID),
		WaybillID: strings.TrimSpace(payload.WaybillID),
		Items:     items,
		Metadata:  payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ExecuteOutbound(c *gin.Context) {
	if h == nil || h.warehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment warehouse service unavailable", nil)
		return
	}
	outboundID := strings.TrimSpace(c.Param("id"))
	if outboundID == "" {
		contracts.ResponseBadRequest(c, "outbound id is required")
		return
	}
	var payload executeOutboundRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.warehouseSvc.ExecuteOutbound(c.Request.Context(), tenantUUID, outboundID, fulfillmentsvc.ExecuteOutboundRequest{
		OperatorID: strings.TrimSpace(payload.OperatorID),
		PackageNo:  payload.PackageNo,
		Metadata:   payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RollbackOutbound(c *gin.Context) {
	if h == nil || h.warehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "fulfillment warehouse service unavailable", nil)
		return
	}
	outboundID := strings.TrimSpace(c.Param("id"))
	if outboundID == "" {
		contracts.ResponseBadRequest(c, "outbound id is required")
		return
	}
	var payload rollbackOutboundRequest
	_ = c.ShouldBindJSON(&payload)
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.warehouseSvc.RollbackOutbound(c.Request.Context(), tenantUUID, outboundID, strings.TrimSpace(payload.Reason))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
