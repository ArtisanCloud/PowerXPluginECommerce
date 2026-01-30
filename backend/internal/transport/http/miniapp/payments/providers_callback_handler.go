package payments

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type ProviderCallbackHandler struct {
	service *paymentssvc.TransactionService
}

func NewProviderCallbackHandler(svc *paymentssvc.TransactionService) *ProviderCallbackHandler {
	return &ProviderCallbackHandler{service: svc}
}

func (h *ProviderCallbackHandler) Callback(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		respondMiniAppPaymentError(c, paymentssvc.ErrPaymentServiceUnavailable)
		return
	}
	providerID := parseUint(strings.TrimSpace(c.Param("id")))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.HandleProviderCallbackRequest(ctx, tenantUUID, providerID, c.Request)
	if err != nil {
		respondMiniAppPaymentError(c, err)
		return
	}
	if resp != nil {
		if err := sendProviderCallbackResponse(c, resp); err != nil {
			respondMiniAppPaymentError(c, err)
		}
		return
	}
	contracts.ResponseSuccess(c, gin.H{"code": "SUCCESS"})
}

func (h *ProviderCallbackHandler) CallbackBySelector(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		respondMiniAppPaymentError(c, paymentssvc.ErrPaymentServiceUnavailable)
		return
	}
	providerType := strings.TrimSpace(c.Param("type"))
	mchID := strings.TrimSpace(c.Param("mchId"))
	appID := strings.TrimSpace(c.Param("appId"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.HandleProviderCallbackRequestBySelector(ctx, tenantUUID, providerType, mchID, appID, c.Request)
	if err != nil {
		respondMiniAppPaymentError(c, err)
		return
	}
	if resp != nil {
		if err := sendProviderCallbackResponse(c, resp); err != nil {
			respondMiniAppPaymentError(c, err)
		}
		return
	}
	contracts.ResponseSuccess(c, gin.H{"code": "SUCCESS"})
}

func parseUint(raw string) uint64 {
	if raw == "" {
		return 0
	}
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
