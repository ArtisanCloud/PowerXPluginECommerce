package after_sales

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	adminsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	caseService     *adminsvc.CaseService
	decisionService *adminsvc.DecisionService
}

type caseActionPayload struct {
	ReasonCode string `json:"reasonCode,omitempty"`
	Note       string `json:"note,omitempty"`
}

func NewHandler(caseService *adminsvc.CaseService, decisionService *adminsvc.DecisionService) *Handler {
	return &Handler{caseService: caseService, decisionService: decisionService}
}

func (h *Handler) ListCases(c *gin.Context) {
	if h == nil || h.caseService == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	resp, err := h.caseService.List(c.Request.Context(), tenantUUID, adminsvc.CaseQuery{
		Status:   strings.TrimSpace(c.Query("status")),
		CaseType: strings.TrimSpace(c.Query("caseType")),
		OrderID:  strings.TrimSpace(c.Query("orderId")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) GetCase(c *gin.Context) {
	if h == nil || h.caseService == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.caseService.Detail(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) AcceptCase(c *gin.Context) {
	h.transitionCase(c, "accept")
}

func (h *Handler) ReviewCase(c *gin.Context) {
	h.transitionCase(c, "review")
}

func (h *Handler) CompleteCase(c *gin.Context) {
	h.transitionCase(c, "complete")
}

func (h *Handler) CloseCase(c *gin.Context) {
	h.transitionCase(c, "close")
}

func (h *Handler) ApproveCase(c *gin.Context) {
	if h == nil || h.decisionService == nil || h.caseService == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales service unavailable", nil)
		return
	}
	var payload caseActionPayload
	if err := c.ShouldBindJSON(&payload); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	if _, err := h.decisionService.Approve(c.Request.Context(), tenantUUID, caseID, adminID, payload.Note); err != nil {
		respondAdminError(c, err)
		return
	}
	detail, err := h.caseService.Detail(c.Request.Context(), tenantUUID, caseID)
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func (h *Handler) RejectCase(c *gin.Context) {
	if h == nil || h.decisionService == nil || h.caseService == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales service unavailable", nil)
		return
	}
	var payload caseActionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	if _, err := h.decisionService.Reject(c.Request.Context(), tenantUUID, caseID, adminID, payload.ReasonCode, payload.Note); err != nil {
		respondAdminError(c, err)
		return
	}
	detail, err := h.caseService.Detail(c.Request.Context(), tenantUUID, caseID)
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func (h *Handler) transitionCase(c *gin.Context, action string) {
	if h == nil || h.caseService == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales service unavailable", nil)
		return
	}
	var payload caseActionPayload
	_ = c.ShouldBindJSON(&payload)
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	if _, err := h.caseService.Transition(c.Request.Context(), tenantUUID, caseID, action, adminID, payload.Note); err != nil {
		respondAdminError(c, err)
		return
	}
	detail, err := h.caseService.Detail(c.Request.Context(), tenantUUID, caseID)
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, detail)
}

func adminUserID(c *gin.Context) (string, bool) {
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return "", false
	}
	return strconv.FormatInt(tc.UserID, 10), true
}

func respondAdminError(c *gin.Context, err error) {
	mapped := contracts.MapAfterSalesError(err)
	contracts.ResponseError(c, mapped.Status, mapped.Code, mapped.Message)
}
