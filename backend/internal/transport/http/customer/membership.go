package customer

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

// ListCustomerEntitlements returns entitlements for an admin-selected customer.
func (h *Handler) ListCustomerEntitlements(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	customerID := strings.TrimSpace(c.Param("id"))
	if customerID == "" {
		contracts.ResponseBadRequest(c, "customer id is required")
		return
	}
	db := h.service.DB()
	if db == nil {
		contracts.ResponseServiceUnavailable(c, "database unavailable", nil)
		return
	}
	var items []membershipModel.Entitlement
	if err := db.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
		Order("created_at DESC").
		Find(&items).Error; err != nil && err != gorm.ErrRecordNotFound {
		contracts.ResponseInternalError(c, err)
		return
	}
	resp := make([]entitlementResponse, 0, len(items))
	for _, it := range items {
		var vf *string
		if it.ValidFrom != nil {
			val := it.ValidFrom.UTC().Format(time.RFC3339)
			vf = &val
		}
		var vt *string
		if it.ValidTo != nil {
			val := it.ValidTo.UTC().Format(time.RFC3339)
			vt = &val
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
	contracts.ResponseSuccess(c, gin.H{"items": resp})
}

// GetCustomerTokenBalances returns token balances for an admin-selected customer.
func (h *Handler) GetCustomerTokenBalances(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	customerID := strings.TrimSpace(c.Param("id"))
	if customerID == "" {
		contracts.ResponseBadRequest(c, "customer id is required")
		return
	}
	db := h.service.DB()
	if db == nil {
		contracts.ResponseServiceUnavailable(c, "database unavailable", nil)
		return
	}
	var rows []membershipModel.TokenAccount
	if err := db.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
		Order("updated_at DESC").
		Find(&rows).Error; err != nil && err != gorm.ErrRecordNotFound {
		contracts.ResponseInternalError(c, err)
		return
	}
	resp := make([]tokenBalanceResponse, 0, len(rows))
	for _, it := range rows {
		resp = append(resp, tokenBalanceResponse{
			TokenCode: strings.TrimSpace(it.TokenCode),
			Balance:   it.Balance,
		})
	}
	contracts.ResponseSuccess(c, gin.H{"items": resp})
}
