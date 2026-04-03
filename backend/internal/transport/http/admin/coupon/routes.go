package coupon

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/coupons namespace.
func RegisterRoutes(router *gin.RouterGroup, _ *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	return router.Group("/coupons", httpmw.EnsureTenant())
}
