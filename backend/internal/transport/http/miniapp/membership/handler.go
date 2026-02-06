package membership

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	customerModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler exposes mini-app membership query endpoints.
type Handler struct {
	deps *app.Deps
}

func NewHandler(deps *app.Deps) *Handler {
	return &Handler{deps: deps}
}

type profileResponse struct {
	CustomerID     string `json:"customerId"`
	CustomerName   string `json:"customerName"`
	MembershipTier string `json:"membershipTier"`
	TierName       string `json:"tierName,omitempty"`
	AvatarURL      string `json:"avatarUrl,omitempty"`
}

type benefitResponse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Items  json.RawMessage `json:"items,omitempty"`
	Status string          `json:"status,omitempty"`
}

type entitlementResponse struct {
	ID          string  `json:"id"`
	ServiceCode string  `json:"serviceCode"`
	Quantity    int64   `json:"quantity"`
	ValidFrom   *string `json:"validFrom,omitempty"`
	ValidTo     *string `json:"validTo,omitempty"`
	StackPolicy string  `json:"stackPolicy"`
	SourceType  string  `json:"sourceType"`
	SourceID    string  `json:"sourceId"`
}

type tokenBalanceResponse struct {
	TokenCode string `json:"tokenCode"`
	Balance   int64  `json:"balance"`
}

func (h *Handler) GetProfile(c *gin.Context) {
	if h == nil || h.deps == nil || h.deps.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "service unavailable"})
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	if strings.TrimSpace(cc.CustomerID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "customer required"})
		return
	}
	var customer customerModel.Customer
	if err := h.deps.DB.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, cc.CustomerID).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	avatarURL := ""
	nickname := ""
	if len(customer.Metadata) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(customer.Metadata, &meta); err == nil {
			if v, ok := meta["avatar_url"].(string); ok {
				avatarURL = strings.TrimSpace(v)
			}
			if v, ok := meta["nickname"].(string); ok {
				nickname = strings.TrimSpace(v)
			}
		}
	}
	tierCode := strings.TrimSpace(customer.MembershipTier)
	tierName := ""
	if tierCode != "" {
		var tier membershipModel.MembershipTier
		if err := h.deps.DB.WithContext(c.Request.Context()).
			Where("tenant_uuid = ? AND LOWER(code) = ?", tenantUUID, strings.ToLower(tierCode)).
			First(&tier).Error; err == nil {
			tierName = strings.TrimSpace(tier.Name)
		}
	}
	resp := profileResponse{
		CustomerID:     strings.TrimSpace(customer.CustomerID),
		CustomerName:   strings.TrimSpace(customer.Name),
		MembershipTier: tierCode,
		TierName:       tierName,
		AvatarURL:      avatarURL,
	}
	if nickname != "" {
		resp.CustomerName = nickname
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *Handler) ListBenefits(c *gin.Context) {
	if h == nil || h.deps == nil || h.deps.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "service unavailable"})
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	if strings.TrimSpace(cc.CustomerID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "customer required"})
		return
	}
	var rows []membershipModel.MembershipBenefit
	if err := h.deps.DB.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND status = ?", tenantUUID, "active").
		Order("created_at DESC").
		Find(&rows).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := make([]benefitResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, benefitResponse{
			ID:     strings.TrimSpace(row.ID),
			Name:   strings.TrimSpace(row.Name),
			Type:   strings.TrimSpace(row.Type),
			Items:  json.RawMessage(row.Items),
			Status: strings.TrimSpace(row.Status),
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": resp}})
}

func (h *Handler) ListEntitlements(c *gin.Context) {
	if h == nil || h.deps == nil || h.deps.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "service unavailable"})
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	if strings.TrimSpace(cc.CustomerID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "customer required"})
		return
	}
	var items []membershipModel.Entitlement
	if err := h.deps.DB.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, cc.CustomerID).
		Order("created_at DESC").
		Find(&items).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := make([]entitlementResponse, 0, len(items))
	for _, it := range items {
		var vf *string
		if it.ValidFrom != nil {
			v := it.ValidFrom.UTC().Format(time.RFC3339)
			vf = &v
		}
		var vt *string
		if it.ValidTo != nil {
			v := it.ValidTo.UTC().Format(time.RFC3339)
			vt = &v
		}
		resp = append(resp, entitlementResponse{
			ID:          it.ID,
			ServiceCode: strings.TrimSpace(it.ServiceCode),
			Quantity:    it.Quantity,
			ValidFrom:   vf,
			ValidTo:     vt,
			StackPolicy: strings.TrimSpace(it.StackPolicy),
			SourceType:  strings.TrimSpace(it.SourceType),
			SourceID:    strings.TrimSpace(it.SourceID),
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": resp}})
}

func (h *Handler) GetTokenBalances(c *gin.Context) {
	if h == nil || h.deps == nil || h.deps.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "service unavailable"})
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	if strings.TrimSpace(cc.CustomerID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "customer required"})
		return
	}
	var rows []membershipModel.TokenAccount
	if err := h.deps.DB.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, cc.CustomerID).
		Order("updated_at DESC").
		Find(&rows).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := make([]tokenBalanceResponse, 0, len(rows))
	for _, it := range rows {
		resp = append(resp, tokenBalanceResponse{
			TokenCode: strings.TrimSpace(it.TokenCode),
			Balance:   it.Balance,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": resp}})
}
