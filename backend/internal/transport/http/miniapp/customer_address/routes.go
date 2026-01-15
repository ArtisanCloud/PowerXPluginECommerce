package customer_address

import (
	addresssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/customer_address"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes exposes /mini-app/me/addresses endpoints for authenticated mini-app users.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	group := router.Group("/me/addresses")
	if deps == nil {
		return group
	}
	handler := NewHandler(addresssvc.NewService(deps))
	group.GET("", handler.List)
	group.POST("", handler.Create)
	group.PATCH("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)
	group.POST("/:id/default", handler.SetDefault)
	return group
}
