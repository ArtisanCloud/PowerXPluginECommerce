package channel_master

import (
	"errors"
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler exposes admin HTTP endpoints for channel master data.
type Handler struct {
	service        *channelservice.Service
	ownerDirectory *channelservice.OwnerDirectory
}

// NewHandler constructs a handler wrapper.
func NewHandler(service *channelservice.Service, ownerDirectory *channelservice.OwnerDirectory) *Handler {
	return &Handler{
		service:        service,
		ownerDirectory: ownerDirectory,
	}
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
		Keyword:   query.Keyword,
		Platform:  query.Platform,
		Status:    query.Status,
		Owner:     query.Owner,
		Region:    query.Region,
		Tags:      query.Tags,
		MinHealth: query.MinHealth,
		MinGMV:    query.MinGMV,
		Page:      query.Page,
		PageSize:  query.PageSize,
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
	var req ChannelUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	summary, err := h.service.CreateDraft(c.Request.Context(), req.ToServiceInput())
	if err != nil {
		respondChannelError(c, err)
		return
	}
	contracts.ResponseCreated(c, summary)
}

// Update mutates draft/rejected channels.
func (h *Handler) Update(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var req ChannelUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	channelID := c.Param("channelId")
	summary, err := h.service.UpdateDraft(c.Request.Context(), channelID, req.ToServiceInput())
	if err != nil {
		respondChannelError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "channel updated")
}

// Submit moves a channel to pending_review state.
func (h *Handler) Submit(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var req SubmitChannelRequest
	_ = c.ShouldBindJSON(&req)
	channelID := c.Param("channelId")
	summary, err := h.service.Submit(c.Request.Context(), channelID, req.ToServiceInput())
	if err != nil {
		respondChannelError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "channel submitted")
}

// Approve processes approve/reject payloads.
func (h *Handler) Approve(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	channelID := c.Param("channelId")
	summary, err := h.service.ProcessApproval(c.Request.Context(), channelID, req.ToServiceInput())
	if err != nil {
		respondChannelError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, summary, "channel approval recorded")
}

// Platforms exposes configured platform + channel-type catalog for the UI.
func (h *Handler) Platforms(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel service unavailable")
		return
	}
	catalog := h.service.PlatformCatalog(c.Request.Context())
	contracts.ResponseSuccess(c, catalog)
}

// Owners exposes IAM member candidates for owner dropdowns.
func (h *Handler) Owners(c *gin.Context) {
	if h.ownerDirectory == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "owner directory unavailable")
		return
	}
	tenantUUID, ok := authx.TenantUUIDFromContext(c.Request.Context())
	if !ok || tenantUUID == "" {
		contracts.ResponseBadRequest(c, "tenant context missing")
		return
	}
	var query OwnerListQuery
	_ = c.ShouldBindQuery(&query)
	owners, err := h.ownerDirectory.Search(c.Request.Context(), tenantUUID, query.Keyword, query.Limit)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": owners})
}

func respondChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		contracts.ResponseNotFound(c, "channel not found")
	case errors.Is(err, channelservice.ErrInvalidStatusTransition):
		contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
	case errors.Is(err, channelservice.ErrDuplicateStore):
		contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeInvalidRequest, err.Error())
	case errors.Is(err, channelservice.ErrInvalidChannelType):
		contracts.ResponseBadRequest(c, err.Error())
	default:
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
	}
}
