package product

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	sellabilitysvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/sellability"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type sellabilityQuery struct {
	Channel string `form:"channel"`
	Locale  string `form:"locale"`
}

// GetSellability returns purchase readiness aggregation for all SKUs under an SPU.
func (h *Handler) GetSellability(c *gin.Context) {
	if h == nil || h.db == nil {
		respondMiniAppError(c, http.StatusServiceUnavailable, errors.New("product service unavailable"))
		return
	}
	spuID := strings.TrimSpace(c.Param("id"))
	if spuID == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("spu id is required"))
		return
	}
	var query sellabilityQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondMiniAppError(c, http.StatusBadRequest, err)
		return
	}
	channel := strings.TrimSpace(query.Channel)
	if channel == "" {
		respondMiniAppError(c, http.StatusBadRequest, errors.New("channel is required"))
		return
	}
	tenantUUID, _ := middleware.TenantUUIDFromContext(c)
	svc := sellabilitysvc.NewService(h.db)
	result, err := svc.Evaluate(c.Request.Context(), tenantUUID, spuID, channel, query.Locale)
	if err != nil {
		respondMiniAppError(c, 0, err)
		return
	}
	contracts.ResponseSuccess(c, result)
}
