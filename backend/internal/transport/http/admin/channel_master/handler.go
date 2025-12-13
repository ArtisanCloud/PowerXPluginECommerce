package channel_master

import (
	"errors"
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler exposes admin HTTP endpoints for channel master data.
type Handler struct {
	service *channelservice.Service
}

// NewHandler constructs a handler wrapper.
func NewHandler(service *channelservice.Service) *Handler {
	return &Handler{service: service}
}

// List returns channel summaries.
func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var query ChannelListQuery
	_ = c.ShouldBindQuery(&query)
	filters := channelrepo.ChannelListFilters{
		Keyword:  query.Keyword,
		Platform: query.Platform,
		Status:   query.Status,
		Owner:    query.Owner,
		Region:   query.Region,
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	page, err := h.service.List(c.Request.Context(), filters)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"items": page.List,
		"meta":  gin.H{"total": page.Total, "page": page.PageIndex, "pageSize": page.PageSize},
	})
}

// Create registers a channel draft.
func (h *Handler) Create(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	summary, err := h.service.CreateDraft(c.Request.Context(), req.ToServiceInput())
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "channel draft created")
}
