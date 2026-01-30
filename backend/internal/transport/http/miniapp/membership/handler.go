package membership

import (
	"net/http"
	"strings"
	"time"

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
