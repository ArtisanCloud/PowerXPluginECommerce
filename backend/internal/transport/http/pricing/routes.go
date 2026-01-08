package pricing

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires tenant-scoped pricing endpoints beneath the given router group.
func RegisterRoutes(router *gin.RouterGroup, _ *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	group := router.Group("/pricing", httpmw.EnsureTenant())
	group.POST("/query", func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "pricing query not implemented"})
	})
	return group
}
