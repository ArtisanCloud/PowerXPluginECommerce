package membership

import (
	"encoding/json"
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

type createTierPayload struct {
	Name   string         `json:"name"`
	Code   string         `json:"code"`
	Status string         `json:"status,omitempty"`
	Rules  map[string]any `json:"rules,omitempty"`
}

type createBenefitPayload struct {
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Items  any    `json:"items,omitempty"`
}

type updateTierStatusPayload struct {
	Status string `json:"status"`
}

type redeemPointsBenefitPayload struct {
	CustomerID string `json:"customerId"`
	BenefitID  string `json:"benefitId"`
	PointsCost int64  `json:"pointsCost"`
	Reason     string `json:"reason,omitempty"`
	SourceID   string `json:"sourceId,omitempty"`
}

type tokenTransactionResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customerId"`
	TokenCode  string `json:"tokenCode"`
	Delta      int64  `json:"delta"`
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	CreatedAt  string `json:"createdAt"`
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

func (h *Handler) CreateTier(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload createTierPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	rulesBytes := []byte("{}")
	if payload.Rules != nil {
		if data, err := json.Marshal(payload.Rules); err == nil {
			rulesBytes = data
		}
	}
	req := membershipsvc.CreateTierRequest{
		Name:    strings.TrimSpace(payload.Name),
		Code:    strings.TrimSpace(payload.Code),
		Status:  strings.TrimSpace(payload.Status),
		Rules:   rulesBytes,
		ActorID: adminID,
	}
	tier, err := h.service.CreateTier(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, tier)
}

func (h *Handler) CreateBenefit(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload createBenefitPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	itemsBytes := []byte("[]")
	if payload.Items != nil {
		if data, err := json.Marshal(payload.Items); err == nil {
			itemsBytes = data
		}
	}
	req := membershipsvc.CreateBenefitRequest{
		Name:    strings.TrimSpace(payload.Name),
		Type:    strings.TrimSpace(payload.Type),
		Status:  strings.TrimSpace(payload.Status),
		Items:   itemsBytes,
		ActorID: adminID,
	}
	benefit, err := h.service.CreateBenefit(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, benefit)
}

func (h *Handler) UpdateTierStatus(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload updateTierStatusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	tier, err := h.service.UpdateTierStatus(c.Request.Context(), tenantUUID, c.Param("id"), membershipsvc.UpdateTierStatusRequest{
		Status:  strings.TrimSpace(payload.Status),
		ActorID: adminID,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, tier)
}

func (h *Handler) DeleteTier(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteTier(c.Request.Context(), tenantUUID, c.Param("id"), adminID); err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"id": c.Param("id")})
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

func (h *Handler) ListTokenTransactions(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var createdFrom *time.Time
	if raw := strings.TrimSpace(c.Query("createdFrom")); raw != "" {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			createdFrom = &ts
		}
	}
	var createdTo *time.Time
	if raw := strings.TrimSpace(c.Query("createdTo")); raw != "" {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			createdTo = &ts
		}
	}
	req := membershipsvc.ListTokenTransactionsRequest{
		CustomerID:  strings.TrimSpace(c.Query("customerId")),
		TokenCode:   strings.TrimSpace(c.Query("tokenCode")),
		SourceType:  strings.TrimSpace(c.Query("sourceType")),
		SourceID:    strings.TrimSpace(c.Query("sourceId")),
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Page:        page,
		PageSize:    pageSize,
	}
	items, total, err := h.service.ListTokenTransactions(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	resp := make([]tokenTransactionResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, tokenTransactionResponse{
			ID:         item.ID,
			CustomerID: strings.TrimSpace(item.CustomerID),
			TokenCode:  strings.TrimSpace(item.TokenCode),
			Delta:      item.Delta,
			SourceType: strings.TrimSpace(item.SourceType),
			SourceID:   strings.TrimSpace(item.SourceID),
			CreatedAt:  item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	contracts.ResponseSuccess(c, gin.H{
		"items":    resp,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

func (h *Handler) RedeemPointsBenefit(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "membership service unavailable", nil)
		return
	}
	var payload redeemPointsBenefitPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	adminID, ok := adminUserID(c)
	if !ok {
		return
	}
	req := membershipsvc.RedeemPointsBenefitRequest{
		CustomerID: strings.TrimSpace(payload.CustomerID),
		BenefitID:  strings.TrimSpace(payload.BenefitID),
		PointsCost: payload.PointsCost,
		Reason:     strings.TrimSpace(payload.Reason),
		ActorID:    adminID,
		SourceID:   strings.TrimSpace(payload.SourceID),
	}
	resp, err := h.service.RedeemPointsBenefit(c.Request.Context(), tenantUUID, req)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func adminUserID(c *gin.Context) (string, bool) {
	tc, ok := authx.GetTenantContext(c)
	if !ok || tc.UserID <= 0 {
		contracts.ResponseUnauthorized(c, "unauthorized")
		return "", false
	}
	return strconv.FormatInt(tc.UserID, 10), true
}
