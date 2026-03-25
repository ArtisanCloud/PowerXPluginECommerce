package reverse

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	HttpMW "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts reverse logistics admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	_ = deps
	return router.Group("/reverse", HttpMW.EnsureTenant())
}
