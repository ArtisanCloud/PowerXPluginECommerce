package pricing

import (
	"errors"
	"net/http"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
)

type VersionsHandler struct {
	domain *pricingsvc.Service
	vs     *pricingsvc.VersionService
}

func NewVersionsHandler(domain *pricingsvc.Service) *VersionsHandler {
	if domain == nil || !domain.Ready() {
		return &VersionsHandler{domain: domain, vs: nil}
	}
	return &VersionsHandler{domain: domain, vs: pricingsvc.NewVersionService(domain.Deps())}
}

func (h *VersionsHandler) Create(c *gin.Context) {
	if h == nil || h.vs == nil || !h.vs.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	if pricebookID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId is required"))
		return
	}
	var req VersionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	v, err := h.vs.CreateDraft(c.Request.Context(), pricingsvc.CreateVersionInput{
		PricebookID:       pricebookID,
		CopyFromVersionID: req.CopyFromVersionID,
		Actor:             actorFromContext(c),
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toVersionDTO(v))
}

func (h *VersionsHandler) Publish(c *gin.Context) {
	if h == nil || h.vs == nil || !h.vs.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	versionID := strings.TrimSpace(c.Param("versionId"))
	if pricebookID == "" || versionID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId/versionId are required"))
		return
	}
	var req VersionPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	v, err := h.vs.Publish(c.Request.Context(), pricingsvc.PublishVersionInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		EffectiveAt: req.EffectiveAt,
		ExpiresAt:   req.ExpiresAt,
		Note:        req.Note,
		Actor:       actorFromContext(c),
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVersionDTO(v))
}

func (h *VersionsHandler) Archive(c *gin.Context) {
	if h == nil || h.vs == nil || !h.vs.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	versionID := strings.TrimSpace(c.Param("versionId"))
	if pricebookID == "" || versionID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId/versionId are required"))
		return
	}
	var req VersionArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	v, err := h.vs.Archive(c.Request.Context(), pricingsvc.ArchiveVersionInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		Note:        req.Note,
		Actor:       actorFromContext(c),
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVersionDTO(v))
}

func toVersionDTO(v *pricingModel.PricebookVersion) VersionDTO {
	if v == nil {
		return VersionDTO{}
	}
	publishedBy := (*string)(nil)
	note := (*string)(nil)
	if strings.TrimSpace(v.PublishedBy) != "" {
		s := v.PublishedBy
		publishedBy = &s
	}
	if strings.TrimSpace(v.Note) != "" {
		s := v.Note
		note = &s
	}
	return VersionDTO{
		ID:          v.ID,
		PricebookID: v.PricebookID,
		Version:     v.Version,
		State:       v.State,
		EffectiveAt: v.EffectiveAt,
		ExpiresAt:   v.ExpiresAt,
		PublishedAt: v.PublishedAt,
		PublishedBy: publishedBy,
		Note:        note,
	}
}
