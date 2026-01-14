package cart

import (
	cartsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/cart"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /mini-app/cart endpoints (protected).
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if rg == nil {
		return nil
	}
	group := rg.Group("/cart")
	if deps == nil || deps.DB == nil {
		return group
	}
	handler := NewHandler(cartsvc.NewService(deps.DB))
	group.GET("", handler.GetCart)
	group.POST("/sync", handler.SyncCart)
	return group
}
