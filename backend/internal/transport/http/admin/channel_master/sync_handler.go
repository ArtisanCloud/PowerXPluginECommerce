package channel_master

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
)

// SyncHandler handles manual sync triggers and listings.
type SyncHandler struct {
	service *channelservice.SyncHistoryService
}

// SyncTriggerRequest payload.
type SyncTriggerRequest struct {
	TriggerType string         `json:"trigger_type"`
	Payload     map[string]any `json:"payload"`
}

// NewSyncHandler builds handler.
func NewSyncHandler(syncSvc *channelservice.SyncHistoryService) *SyncHandler {
	return &SyncHandler{service: syncSvc}
}

// Trigger enqueues a manual sync.
func (h *SyncHandler) Trigger(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "sync service unavailable")
		return
	}
	var req SyncTriggerRequest
	_ = c.ShouldBindJSON(&req)
	entry, err := h.service.TriggerManual(c.Request.Context(), c.Param("channelId"), channelservice.SyncTriggerInput{
		TriggerType: req.TriggerType,
		Payload:     req.Payload,
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"historyId": entry.ID,
		"status":    entry.Result,
	})
}

// History returns latest executions.
func (h *SyncHandler) History(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "sync service unavailable")
		return
	}
	items, err := h.service.ListRecent(c.Request.Context(), c.Param("channelId"), 10)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": mapSyncDTO(items)})
}
