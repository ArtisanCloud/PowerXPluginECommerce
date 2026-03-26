package logistics

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListWaybillETA(c *gin.Context) {
	if h == nil || h.etaSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics eta service unavailable", nil)
		return
	}
	rawIDs := strings.TrimSpace(c.Query("waybill_ids"))
	ids := make([]string, 0)
	if rawIDs != "" {
		for _, v := range strings.Split(rawIDs, ",") {
			if id := strings.TrimSpace(v); id != "" {
				ids = append(ids, id)
			}
		}
	}
	force := false
	if v := strings.TrimSpace(c.Query("force_recompute")); v != "" {
		force = v == "1" || strings.EqualFold(v, "true")
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.etaSvc.ListByWaybillIDs(c.Request.Context(), tenantUUID, ids, logisticssvc.ETAQuery{
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		Timezone:        strings.TrimSpace(c.Query("timezone")),
		ForceRecompute:  force,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) GetWaybillETA(c *gin.Context) {
	if h == nil || h.etaSvc == nil {
		contracts.ResponseServiceUnavailable(c, "logistics eta service unavailable", nil)
		return
	}
	waybillID := strings.TrimSpace(c.Param("waybill_id"))
	if waybillID == "" {
		contracts.ResponseBadRequest(c, "waybill_id is required")
		return
	}
	force := false
	if v := strings.TrimSpace(c.Query("force_recompute")); v != "" {
		force = v == "1" || strings.EqualFold(v, "true")
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.etaSvc.GetByWaybillID(c.Request.Context(), tenantUUID, waybillID, logisticssvc.ETAQuery{
		DestinationZone: strings.TrimSpace(c.Query("destination_zone")),
		Timezone:        strings.TrimSpace(c.Query("timezone")),
		ForceRecompute:  force,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
