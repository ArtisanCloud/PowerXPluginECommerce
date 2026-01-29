package payments

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	service *paymentssvc.TransactionService
}

func NewTransactionHandler(svc *paymentssvc.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: svc}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		respondMiniAppPaymentError(c, paymentssvc.ErrPaymentServiceUnavailable)
		return
	}
	var req paymentssvc.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondMiniAppPaymentError(c, paymentssvc.ErrInvalidArgument)
		return
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		req.IdempotencyKey = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if strings.TrimSpace(tenantUUID) == "" {
		if cc, ok := authx.GetCustomerContext(c); ok {
			tenantUUID = strings.TrimSpace(cc.TenantUUID)
		}
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.CreateTransaction(ctx, tenantUUID, req)
	if err != nil {
		respondMiniAppPaymentError(c, err)
		return
	}
	contracts.ResponseCreated(c, resp)
}

func (h *TransactionHandler) GetStatus(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		respondMiniAppPaymentError(c, paymentssvc.ErrPaymentServiceUnavailable)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	if strings.TrimSpace(tenantUUID) == "" {
		if cc, ok := authx.GetCustomerContext(c); ok {
			tenantUUID = strings.TrimSpace(cc.TenantUUID)
		}
	}
	transactionID := strings.TrimSpace(c.Param("id"))
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.GetTransactionStatus(ctx, tenantUUID, transactionID)
	if err != nil {
		respondMiniAppPaymentError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}
