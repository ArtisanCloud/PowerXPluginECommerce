package channel_master

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CredentialHandler wires credential endpoints.
type CredentialHandler struct {
	service *channelservice.CredentialService
}

// NewCredentialHandler constructs handler.
func NewCredentialHandler(service *channelservice.CredentialService) *CredentialHandler {
	return &CredentialHandler{service: service}
}

// List returns credential metadata per channel.
func (h *CredentialHandler) List(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "credential service unavailable")
		return
	}
	channelID := c.Param("channelId")
	creds, err := h.service.List(c.Request.Context(), channelID)
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": creds})
}

// Upsert stores encrypted credential payload.
func (h *CredentialHandler) Upsert(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "credential service unavailable")
		return
	}
	var req CredentialUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	dto, err := h.service.Upsert(c.Request.Context(), c.Param("channelId"), channelservice.CredentialUpsertInput{
		Type:          req.Type,
		Payload:       req.Payload,
		Scope:         req.Scope,
		ExpiresAt:     req.ExpiresAt,
		Metadata:      req.Metadata,
		AttachmentURL: req.AttachmentURL,
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, dto, "credential saved")
}

// Test updates credential test result.
func (h *CredentialHandler) Test(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "credential service unavailable")
		return
	}
	var req CredentialTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	dto, err := h.service.RecordTestResult(c.Request.Context(), c.Param("channelId"), channelservice.CredentialTestInput{
		Type:      req.Type,
		Result:    req.Result,
		Succeeded: req.Succeeded,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		contracts.ResponseError(c, status, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, dto, "credential test result recorded")
}
