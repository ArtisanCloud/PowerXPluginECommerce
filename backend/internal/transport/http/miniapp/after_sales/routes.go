package after_sales

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts after-sales mini-app route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	_ = deps
	return router.Group("/after-sales")
}
