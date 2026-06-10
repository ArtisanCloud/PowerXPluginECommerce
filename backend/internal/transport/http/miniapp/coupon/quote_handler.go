package coupon

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *couponsvc.QuoteService
}

func NewHandler(deps *app.Deps) *Handler {
	return &Handler{svc: couponsvc.NewQuoteService(deps)}
}

func (h *Handler) Quote(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		contracts.ResponseServiceUnavailable(c, "coupon quote service unavailable", nil)
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
	items := make([]couponsvc.QuoteItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, couponsvc.QuoteItemInput{
			LineID:         strings.TrimSpace(item.LineID),
			SKUID:          strings.TrimSpace(item.SKUID),
			Qty:            item.Qty,
			UnitPriceMinor: item.UnitPriceMinor,
		})
	}
	out, err := h.svc.Quote(c.Request.Context(), couponsvc.QuoteInput{
		TenantUUID: strings.TrimSpace(tenantUUID),
		UserID:     strings.TrimSpace(req.UserID),
		Channel:    strings.TrimSpace(req.Channel),
		CouponIDs:  req.CouponIDs,
		Items:      items,
		Currency:   strings.TrimSpace(req.Currency),
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
