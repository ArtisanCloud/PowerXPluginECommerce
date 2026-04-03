package coupon

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /v1/coupons namespace.
func RegisterRoutes(router *gin.RouterGroup, _ *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	group := router.Group("/coupons", httpmw.EnsureTenant())
	group.POST("/quote", Quote)
	return group
}
