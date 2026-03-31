package after_sales

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	adminsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *adminsvc.DashboardService
}

func NewDashboardHandler(service *adminsvc.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetSnapshot(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales dashboard service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.Snapshot(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
