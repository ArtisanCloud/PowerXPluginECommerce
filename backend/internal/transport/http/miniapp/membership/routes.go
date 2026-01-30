package membership

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires membership endpoints for mini-app.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil {
		return
	}
	h := NewHandler(deps)
	group := rg.Group("/membership")
	group.GET("/entitlements", h.ListEntitlements)
	group.GET("/tokens", h.GetTokenBalances)
}
