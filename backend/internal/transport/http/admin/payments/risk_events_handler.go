package payments

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	paymentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type RiskEventHandler struct {
	service *paymentsvc.RiskEventService
}

func NewRiskEventHandler(svc *paymentsvc.RiskEventService) *RiskEventHandler {
	return &RiskEventHandler{service: svc}
}

func (h *RiskEventHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "risk event service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var txID *uint64
	if raw := strings.TrimSpace(c.Query("transactionId")); raw != "" {
		if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
			txID = &parsed
		}
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListRiskEvents(ctx, tenantUUID, adminID, txID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
