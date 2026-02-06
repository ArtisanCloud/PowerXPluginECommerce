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
	rg.GET("/benefits", handler.ListBenefits)
	rg.POST("/entitlements/grant", handler.GrantEntitlement)
	rg.POST("/entitlements/revoke", handler.RevokeEntitlement)
	rg.POST("/tokens/adjust", handler.AdjustToken)
	return rg
}
