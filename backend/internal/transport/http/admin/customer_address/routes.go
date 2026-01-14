package customer_address

import (
	addresssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/customer_address"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/customers/:customerId/addresses namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/customers/:id/addresses", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(addresssvc.NewService(deps))
	rg.GET("", handler.List)
	rg.POST("", handler.Create)
	rg.PATCH("/:addressId", handler.Update)
	rg.DELETE("/:addressId", handler.Delete)
	rg.POST("/:addressId/default", handler.SetDefault)
	return rg
}
