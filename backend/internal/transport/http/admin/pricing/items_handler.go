package pricing

import (
	"errors"
	"net/http"
	"strings"

	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
)

type ItemsHandler struct {
	items *pricingsvc.ItemService
}

func NewItemsHandler(domain *pricingsvc.Service) *ItemsHandler {
	if domain == nil || !domain.Ready() {
		return &ItemsHandler{items: nil}
	}
	return &ItemsHandler{items: pricingsvc.NewItemService(domain.Deps())}
}

func (h *ItemsHandler) Upsert(c *gin.Context) {
	if h == nil || h.items == nil || !h.items.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	versionID := strings.TrimSpace(c.Param("versionId"))
	if pricebookID == "" || versionID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId/versionId are required"))
		return
	}

	var req ItemsUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	items := make([]pricingsvc.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, pricingsvc.ItemInput{
			SKUID:           it.SKUID,
			BaseAmountMinor: it.BaseAmountMinor,
			SaleAmountMinor: it.SaleAmountMinor,
			MsrpAmountMinor: it.MsrpAmountMinor,
			CostAmountMinor: it.CostAmountMinor,
			MinAmountMinor:  it.MinAmountMinor,
			MaxAmountMinor:  it.MaxAmountMinor,
			TaxIncluded:     it.TaxIncluded,
			Meta:            it.Meta,
		})
	}

	res, err := h.items.UpsertItems(c.Request.Context(), pricingsvc.UpsertItemsInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		Items:       items,
		Actor:       actorFromContext(c),
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ItemsUpsertResponse{
		Upserted: res.Upserted,
		Skipped:  res.Skipped,
	})
}
