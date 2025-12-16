package channel_master

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
)

// TaskNoteHandler exposes task/note endpoints.
type TaskNoteHandler struct {
	service *channelservice.TaskNoteService
}

// NewTaskNoteHandler builds handler.
func NewTaskNoteHandler(taskSvc *channelservice.TaskNoteService) *TaskNoteHandler {
	return &TaskNoteHandler{service: taskSvc}
}

// ListTasks returns existing links.
func (h *TaskNoteHandler) ListTasks(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "tasks unavailable")
		return
	}
	items, err := h.service.ListTasks(c.Request.Context(), c.Param("channelId"))
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": mapTaskDTO(items)})
}

// LinkTask associates channel to remediation task.
func (h *TaskNoteHandler) LinkTask(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "tasks unavailable")
		return
	}
	var req channelservice.TaskLinkInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	link, err := h.service.LinkTask(c.Request.Context(), c.Param("channelId"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	dto := mapTaskDTO([]*channelmodel.ChannelTaskLink{link})
	if len(dto) == 0 {
		contracts.ResponseSuccess(c, gin.H{})
		return
	}
	contracts.ResponseSuccess(c, dto[0])
}

// UpdateTask updates status/note.
func (h *TaskNoteHandler) UpdateTask(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "tasks unavailable")
		return
	}
	var req channelservice.TaskLinkInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	if err := h.service.UpdateTask(c.Request.Context(), c.Param("taskLinkId"), req.Status, req.Note); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"updated": true})
}

// RemoveTask deletes task link.
func (h *TaskNoteHandler) RemoveTask(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "tasks unavailable")
		return
	}
	if err := h.service.RemoveTask(c.Request.Context(), c.Param("taskLinkId")); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"deleted": true})
}

// ListNotes returns notes.
func (h *TaskNoteHandler) ListNotes(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "notes unavailable")
		return
	}
	items, err := h.service.ListNotes(c.Request.Context(), c.Param("channelId"), 20)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": mapNoteDTO(items)})
}

// AddNote creates new note.
func (h *TaskNoteHandler) AddNote(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "notes unavailable")
		return
	}
	var req channelservice.NoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	note, err := h.service.AddNote(c.Request.Context(), c.Param("channelId"), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	dto := mapNoteDTO([]*channelmodel.ChannelNote{note})
	if len(dto) == 0 {
		contracts.ResponseSuccess(c, gin.H{})
		return
	}
	contracts.ResponseSuccess(c, dto[0])
}
