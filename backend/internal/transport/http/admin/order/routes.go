package order

import (
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/order"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/orders namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/orders", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(ordersvc.NewService(deps))
	rg.GET("", handler.ListOrders)
	rg.GET("/:id", handler.GetOrder)
	rg.POST("", handler.CreateOrder)
	rg.POST("/:id/cancel", handler.CancelOrder)
	return rg
}
