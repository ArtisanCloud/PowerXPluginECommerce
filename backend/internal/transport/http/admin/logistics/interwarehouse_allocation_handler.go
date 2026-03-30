package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListInterwarehouseAllocationCandidates(c *gin.Context) {
	if h == nil || h.interwarehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "interwarehouse allocation service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.interwarehouseSvc.List(c.Request.Context(), tenantUUID, buildInterwarehouseSuggestionQuery(c))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) SuggestInterwarehouseAllocation(c *gin.Context) {
	if h == nil || h.interwarehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "interwarehouse allocation service unavailable", nil)
		return
	}
	var payload suggestInterwarehouseAllocationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.interwarehouseSvc.Suggest(c.Request.Context(), tenantUUID, logisticssvc.SuggestInterwarehouseAllocationRequest{
		RequestKey:        strings.TrimSpace(payload.RequestKey),
		WaybillID:         strings.TrimSpace(payload.WaybillID),
		OrderID:           strings.TrimSpace(payload.OrderID),
		CarrierID:         strings.TrimSpace(payload.CarrierID),
		SourceWarehouseID: strings.TrimSpace(payload.SourceWarehouseID),
		DestinationZone:   strings.TrimSpace(payload.DestinationZone),
		RequiredQty:       payload.RequiredQty,
		OperatorID:        strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ConfirmInterwarehouseAllocation(c *gin.Context) {
	if h == nil || h.interwarehouseSvc == nil {
		contracts.ResponseServiceUnavailable(c, "interwarehouse allocation service unavailable", nil)
		return
	}
	var payload confirmInterwarehouseAllocationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.interwarehouseSvc.Confirm(c.Request.Context(), tenantUUID, logisticssvc.ConfirmInterwarehouseAllocationRequest{
		CandidateID:       strings.TrimSpace(payload.CandidateID),
		RequestKey:        strings.TrimSpace(payload.RequestKey),
		TargetWarehouseID: strings.TrimSpace(payload.TargetWarehouseID),
		OperatorID:        strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}

func buildInterwarehouseSuggestionQuery(c *gin.Context) logisticssvc.InterwarehouseSuggestionQuery {
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	return logisticssvc.InterwarehouseSuggestionQuery{
		RequestKey:        strings.TrimSpace(c.Query("request_key")),
		CarrierID:         strings.TrimSpace(c.Query("carrier_id")),
		SourceWarehouseID: strings.TrimSpace(c.Query("source_warehouse_id")),
		TargetWarehouseID: strings.TrimSpace(c.Query("target_warehouse_id")),
		Status:            strings.TrimSpace(c.Query("status")),
		Limit:             limit,
	}
}
