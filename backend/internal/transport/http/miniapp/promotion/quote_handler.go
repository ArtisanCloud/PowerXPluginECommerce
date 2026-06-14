package promotion

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *promotionsvc.QuoteService
}

func NewHandler(deps *app.Deps) *Handler {
	return &Handler{svc: promotionsvc.NewQuoteService(deps)}
}

func (h *Handler) Quote(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		contracts.ResponseServiceUnavailable(c, "promotion quote service unavailable", nil)
		return
	}
	var req quoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	items := make([]promotionsvc.QuoteItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, promotionsvc.QuoteItemInput{
			LineID: strings.TrimSpace(item.LineID), SKUID: strings.TrimSpace(item.SKUID),
			Qty: item.Qty, UnitPriceMinor: item.UnitPriceMinor,
		})
	}
	var submittedAt time.Time
	if strings.TrimSpace(req.SubmittedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.SubmittedAt))
		if err != nil {
			contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "submitted_at must be RFC3339")
			return
		}
		submittedAt = parsed
	}
	out, err := h.svc.Quote(c.Request.Context(), promotionsvc.QuoteInput{
		TenantUUID: strings.TrimSpace(tenantUUID), UserID: strings.TrimSpace(req.UserID),
		Channel: strings.TrimSpace(req.Channel), Currency: strings.TrimSpace(req.Currency),
		SubmittedAt: submittedAt, Items: items,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseInternalError(c, errors.New("unknown error"))
		return
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "unknown error"
	}
	switch {
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"):
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, msg)
	default:
		contracts.ResponseInternalError(c, err)
	}
}
