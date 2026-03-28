package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListReconciliationBatches(c *gin.Context) {
	if h == nil || h.reconcileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reconciliation service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.reconcileSvc.ListBatches(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("carrier_id")),
		strings.TrimSpace(c.Query("status")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) CreateReconciliationBatch(c *gin.Context) {
	if h == nil || h.reconcileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reconciliation service unavailable", nil)
		return
	}
	var payload createReconciliationBatchRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	records := make([]logisticssvc.ReconciliationInputRecord, 0, len(payload.Records))
	for _, row := range payload.Records {
		records = append(records, logisticssvc.ReconciliationInputRecord{
			WaybillID:     strings.TrimSpace(row.WaybillID),
			WaybillNo:     strings.TrimSpace(row.WaybillNo),
			CarrierID:     strings.TrimSpace(row.CarrierID),
			BillAmount:    row.BillAmount,
			BankAmount:    row.BankAmount,
			InvoiceAmount: row.InvoiceAmount,
		})
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.reconcileSvc.CreateBatch(c.Request.Context(), tenantUUID, logisticssvc.CreateReconciliationBatchRequest{
		CarrierID: strings.TrimSpace(payload.CarrierID),
		From:      strings.TrimSpace(payload.From),
		To:        strings.TrimSpace(payload.To),
		Records:   records,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListReconciliationRecords(c *gin.Context) {
	if h == nil || h.reconcileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reconciliation service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.reconcileSvc.ListRecords(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("batch_id")),
		strings.TrimSpace(c.Query("status")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ListReconciliationCases(c *gin.Context) {
	if h == nil || h.reconcileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reconciliation service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.reconcileSvc.ListCases(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("batch_id")),
		strings.TrimSpace(c.Query("status")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) HandleReconciliationCase(c *gin.Context) {
	if h == nil || h.reconcileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "reconciliation service unavailable", nil)
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	if caseID == "" {
		contracts.ResponseBadRequest(c, "case id is required")
		return
	}
	var payload handleReconciliationCaseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.reconcileSvc.HandleCase(c.Request.Context(), tenantUUID, caseID, logisticssvc.HandleReconciliationCaseRequest{
		Action:     strings.TrimSpace(payload.Action),
		OperatorID: strings.TrimSpace(payload.OperatorID),
		Note:       strings.TrimSpace(payload.Note),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
