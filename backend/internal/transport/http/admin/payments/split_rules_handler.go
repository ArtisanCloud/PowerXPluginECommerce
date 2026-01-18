package payments

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	paymentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type SplitRuleHandler struct {
	service *paymentsvc.SplitRuleService
}

func NewSplitRuleHandler(svc *paymentsvc.SplitRuleService) *SplitRuleHandler {
	return &SplitRuleHandler{service: svc}
}

func (h *SplitRuleHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "split rule service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListSplitRules(ctx, tenantUUID, adminID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
