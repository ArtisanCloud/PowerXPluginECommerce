package order

import (
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/order"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes exposes /mini-app/orders endpoints for authenticated mini-app users.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	handler := NewHandler(ordersvc.NewService(deps))
	group := router.Group("/orders")
	group.POST("", handler.CreateOrder)
	group.GET("", handler.ListOrders)
	group.GET("/:id", handler.GetOrder)
	return group
}
