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

type TransactionHandler struct {
	service *paymentsvc.TransactionService
}

func NewTransactionHandler(svc *paymentsvc.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: svc}
}

func (h *TransactionHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "transaction service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var providerID uint64
	if v := strings.TrimSpace(c.Query("providerId")); v != "" {
		if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
			providerID = parsed
		}
	}
	from, err := paymentsvc.ParseTimeQuery(c.Query("from"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid from time")
		return
	}
	to, err := paymentsvc.ParseTimeQuery(c.Query("to"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid to time")
		return
	}
	filter := paymentsvc.TransactionListFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		ProviderID: providerID,
		OrderID:    strings.TrimSpace(c.Query("orderId")),
		OrderNo:    strings.TrimSpace(c.Query("orderNo")),
		From:       from,
		To:         to,
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.ListTransactions(ctx, tenantUUID, adminID, filter)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *TransactionHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "transaction service unavailable")
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := requireAdminUser(c)
	if !ok {
		return
	}
	idRaw := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "invalid transaction id")
		return
	}
	ctx := authx.ContextWithRequestID(c.Request.Context(), requestIDFromRequest(c))
	resp, err := h.service.GetTransaction(ctx, tenantUUID, adminID, id)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}
