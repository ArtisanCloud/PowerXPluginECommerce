package promotion

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type CampaignHandler struct {
	svc *promotionsvc.CampaignService
}

func NewCampaignHandler(svc *promotionsvc.CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

func (h *CampaignHandler) List(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		contracts.ResponseServiceUnavailable(c, "promotion service unavailable", nil)
		return
	}
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	out, err := h.svc.List(c.Request.Context(), tenantUUID, promotionsvc.CampaignListFilter{
		Keyword: c.Query("keyword"), PromotionType: c.Query("promotion_type"), Status: c.Query("status"), Channel: c.Query("channel"),
		Page: parsePage(c.Query("page"), 1), PageSize: parsePage(c.Query("page_size"), 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *CampaignHandler) Create(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		contracts.ResponseServiceUnavailable(c, "promotion service unavailable", nil)
		return
	}
	actor, ok := requireAdminUser(c)
	if !ok {
		return
	}
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	var req campaignCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), promotionsvc.CampaignCreateInput{
		TenantUUID: tenantUUID, Code: req.Code, Name: req.Name, Description: req.Description, PromotionType: req.PromotionType,
		ConditionRule: req.ConditionRule, ScopeRule: req.ScopeRule, ActionRule: req.ActionRule, StackingRule: req.StackingRule,
		ValidFrom: req.ValidFrom, ValidTo: req.ValidTo, SaveAction: req.SaveAction, Actor: actor, RequestID: requestIDFromRequest(c),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *CampaignHandler) Update(c *gin.Context) {
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	actor, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var req campaignUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), tenantUUID, c.Param("id"), promotionsvc.CampaignUpdateInput{
		Name: req.Name, Description: req.Description, ConditionRule: req.ConditionRule, ScopeRule: req.ScopeRule,
		ActionRule: req.ActionRule, StackingRule: req.StackingRule, ValidFrom: req.ValidFrom, ValidTo: req.ValidTo,
		Actor: actor, RequestID: requestIDFromRequest(c),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *CampaignHandler) Activate(c *gin.Context) { h.changeStatus(c, "activate") }
func (h *CampaignHandler) Pause(c *gin.Context)    { h.changeStatus(c, "pause") }

func (h *CampaignHandler) changeStatus(c *gin.Context, action string) {
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	actor, ok := requireAdminUser(c)
	if !ok {
		return
	}
	var out any
	var err error
	if action == "activate" {
		out, err = h.svc.Activate(c.Request.Context(), tenantUUID, c.Param("id"), actor, requestIDFromRequest(c))
	} else {
		out, err = h.svc.Pause(c.Request.Context(), tenantUUID, c.Param("id"), actor, requestIDFromRequest(c))
	}
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func (h *CampaignHandler) Clone(c *gin.Context) {
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	actor, ok := requireAdminUser(c)
	if !ok {
		return
	}
	out, err := h.svc.Clone(c.Request.Context(), tenantUUID, c.Param("id"), actor, requestIDFromRequest(c))
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, out)
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseInternalError(c, errors.New("unknown error"))
		return
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "unknown error"
	}
	switch {
	case errors.Is(err, promotionsvc.ErrPromotionNotFound):
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, msg)
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"), strings.Contains(msg, "must"):
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, msg)
	default:
		contracts.ResponseInternalError(c, err)
	}
}
