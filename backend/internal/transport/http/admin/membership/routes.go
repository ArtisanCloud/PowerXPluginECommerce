package membership

import (
	membershipsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/membership"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/membership namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/membership", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(membershipsvc.NewService(deps))
	rg.GET("/tiers", handler.ListTiers)
	rg.POST("/tiers", handler.CreateTier)
	rg.PATCH("/tiers/:id/status", handler.UpdateTierStatus)
	rg.DELETE("/tiers/:id", handler.DeleteTier)
	rg.GET("/benefits", handler.ListBenefits)
	rg.POST("/benefits", handler.CreateBenefit)
	rg.POST("/entitlements/grant", handler.GrantEntitlement)
	rg.POST("/entitlements/revoke", handler.RevokeEntitlement)
	rg.POST("/tokens/adjust", handler.AdjustToken)
	rg.GET("/tokens/transactions", handler.ListTokenTransactions)
	rg.POST("/points/redeem", handler.RedeemPointsBenefit)
	return rg
}
