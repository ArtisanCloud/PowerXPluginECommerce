package pricing

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/pricing namespace. Concrete endpoints are attached in later phases.
func RegisterRoutes(router *gin.RouterGroup, _ *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	return router.Group("/pricing")
}
