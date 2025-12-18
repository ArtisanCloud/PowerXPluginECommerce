package channel_master

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
)

// StrategyHandler exposes HTTP endpoints for config/team updates.
type StrategyHandler struct {
	service *channelservice.StrategyService
}

// NewStrategyHandler constructs handler.
func NewStrategyHandler(service *channelservice.StrategyService) *StrategyHandler {
	return &StrategyHandler{service: service}
}

// Get returns current strategy + team snapshot.
func (h *StrategyHandler) Get(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "strategy service unavailable")
		return
	}
	snapshot, err := h.service.GetStrategy(c.Request.Context(), c.Param("channelId"))
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, mapStrategySnapshot(snapshot))
}

// Upsert updates strategy/team config.
func (h *StrategyHandler) Upsert(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "strategy service unavailable")
		return
	}
	var req StrategyUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	snapshot, err := h.service.UpsertStrategy(c.Request.Context(), c.Param("channelId"), req.ToServiceInput())
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, mapStrategySnapshot(snapshot), "strategy updated")
}
