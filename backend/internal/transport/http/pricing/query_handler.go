package pricing

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	svc *pricingsvc.QueryService
}

func NewQueryHandler(deps *app.Deps) *QueryHandler {
	return &QueryHandler{svc: pricingsvc.NewQueryService(deps)}
}

func (h *QueryHandler) Query(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	var req PricingQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}
	req.SKUID = strings.TrimSpace(req.SKUID)
	req.Currency = strings.TrimSpace(req.Currency)
	if req.SKUID == "" || req.Currency == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, fmt.Errorf("sku_id and currency are required"))
		return
	}

	out, err := h.svc.Query(c.Request.Context(), pricingsvc.QueryInput{
		SKUID:           req.SKUID,
		Currency:        req.Currency,
		ChannelID:       req.ChannelID,
		CustomerGroupID: req.CustomerGroupID,
		SupplierID:      req.SupplierID,
		AsOf:            req.AsOf,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	resp := PricingQueryResponse{
		Currency:    out.Currency,
		Priced:      out.Priced,
		SourceField: out.SourceField,
		Fields:      out.Fields,
		Trace:       PricingTrace{Fallback: out.Trace.Fallback, NoPriceReason: out.Trace.NoPriceReason},
	}
	if out.Matched != nil {
		resp.Matched = &PricingMatched{
			PricebookID:   out.Matched.PricebookID,
			PricebookCode: out.Matched.PricebookCode,
			VersionID:     out.Matched.VersionID,
			Version:       out.Matched.Version,
			SKUID:         out.Matched.SKUID,
		}
	}
	if out.Priced && out.AmountMinor != nil {
		resp.Price = &Money{
			AmountMinor: *out.AmountMinor,
			Amount:      formatAmount(*out.AmountMinor),
		}
	}

	c.JSON(http.StatusOK, resp)
}

func formatAmount(minor int64) string {
	if minor == 0 {
		return "0.00"
	}
	v := float64(minor) / 100.0
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", v)
}
