package membership

import (
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	membershipsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/membership"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *membershipsvc.Service
}

func NewHandler(svc *membershipsvc.Service) *Handler {
	return &Handler{service: svc}
}

type grantEntitlementPayload struct {
	CustomerID  string `json:"customerId"`
	ServiceCode string `json:"serviceCode"`
	Quantity    int64  `json:"quantity"`
	ValidFrom   string `json:"validFrom,omitempty"`
	ValidTo     string `json:"validTo,omitempty"`
	StackPolicy string `json:"stackPolicy,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type revokeEntitlementPayload struct {
	EntitlementID string `json:"entitlementId"`
	Reason        string `json:"reason,omitempty"`
}

type adjustTokenPayload struct {
	CustomerID string `json:"customerId"`
	TokenCode  string `json:"tokenCode"`
	Delta      int64  `json:"delta"`
	Reason     string `json:"reason,omitempty"`
}

func (h *Handler) ListTiers(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.ListTiers(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": resp})
}

func (h *Handler) ListBenefits(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.ListBenefits(c.Request.Context(), tenantUUID)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": resp})
}

func (h *Handler) GrantEntitlement(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload grantEntitlementPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	req := membershipsvc.GrantEntitlementRequest{
		CustomerID:  strings.TrimSpace(payload.CustomerID),
		ServiceCode: strings.TrimSpace(payload.ServiceCode),
		Quantity:    payload.Quantity,
		StackPolicy: strings.TrimSpace(payload.StackPolicy),
		Reason:      strings.TrimSpace(payload.Reason),
		ActorID:     adminID,
	}
	if payload.ValidFrom != "" {
		if ts, err := time.Parse(time.RFC3339, payload.ValidFrom); err == nil {
			req.ValidFrom = &ts
		}
	}
	if payload.ValidTo != "" {
		if ts, err := time.Parse(time.RFC3339, payload.ValidTo); err == nil {
			req.ValidTo = &ts
		}
	}
	ent, err := h.service.GrantEntitlement(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, ent)
}

func (h *Handler) RevokeEntitlement(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload revokeEntitlementPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	req := membershipsvc.RevokeEntitlementRequest{
		EntitlementID: strings.TrimSpace(payload.EntitlementID),
		Reason:        strings.TrimSpace(payload.Reason),
		ActorID:       adminID,
	}
	if err := h.service.RevokeEntitlement(c.Request.Context(), tenantUUID, req); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"id": req.EntitlementID})
}

func (h *Handler) AdjustToken(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload adjustTokenPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	req := membershipsvc.AdjustTokenRequest{
		CustomerID: strings.TrimSpace(payload.CustomerID),
		TokenCode:  strings.TrimSpace(payload.TokenCode),
		Delta:      payload.Delta,
		Reason:     strings.TrimSpace(payload.Reason),
		ActorID:    adminID,
	}
	account, err := h.service.AdjustToken(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, account)
}

func adminUserID(c *gin.Context) (string, bool) {
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return "", false
	}
	return strconv.FormatInt(tc.UserID, 10), true
}
