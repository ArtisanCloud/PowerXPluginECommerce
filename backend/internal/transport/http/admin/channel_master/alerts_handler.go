package channel_master

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// AlertsHandler exposes CRUD for alerts.
type AlertsHandler struct {
	repo *channelrepo.ChannelAlertRepository
}

// UpdateAlertRequest payload.
type UpdateAlertRequest struct {
	Status       string `json:"status"`
	AssigneeUUID string `json:"assignee_uuid"`
	TaskID       string `json:"task_id"`
	Note         string `json:"note"`
}

// NewAlertsHandler builds handler.
func NewAlertsHandler(deps *app.Deps) *AlertsHandler {
	if deps == nil || deps.DB == nil {
		return &AlertsHandler{}
	}
	return &AlertsHandler{repo: channelrepo.NewChannelAlertRepository(deps.DB)}
}

// List returns channel alerts.
func (h *AlertsHandler) List(c *gin.Context) {
	if h.repo == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "alerts unavailable")
		return
	}
	alerts, err := h.repo.ListByChannel(c.Request.Context(), c.Param("channelId"), 50)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": mapAlertDTO(alerts)})
}

// Update mutates alert status/assignee.
func (h *AlertsHandler) Update(c *gin.Context) {
	if h.repo == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "alerts unavailable")
		return
	}
	var req UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.AssigneeUUID != "" {
		updates["assignee_uuid"] = req.AssigneeUUID
	}
	if req.TaskID != "" {
		updates["task_id"] = req.TaskID
	}
	if req.Note != "" {
		updates["description"] = req.Note
	}
	if err := h.repo.UpdateAttrs(c.Request.Context(), c.Param("alertId"), updates); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"updated": true})
}
