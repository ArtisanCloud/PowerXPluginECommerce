package pricing

import (
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/pricing namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/pricing", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}

	domain := pricingsvc.NewService(deps)
	pricebooks := NewPricebooksHandler(domain)
	versions := NewVersionsHandler(domain)
	items := NewItemsHandler(domain)

	rg.GET("/pricebooks", pricebooks.List)
	rg.POST("/pricebooks", pricebooks.Create)
	rg.PATCH("/pricebooks/:pricebookId", pricebooks.Update)
	rg.DELETE("/pricebooks/:pricebookId", pricebooks.Delete)

	rg.GET("/pricebooks/:pricebookId/versions", versions.List)
	rg.POST("/pricebooks/:pricebookId/versions", versions.Create)
	rg.POST("/pricebooks/:pricebookId/versions/:versionId/publish", versions.Publish)
	rg.POST("/pricebooks/:pricebookId/versions/:versionId/archive", versions.Archive)
	rg.GET("/pricebooks/:pricebookId/versions/:versionId/items", items.List)
	rg.PUT("/pricebooks/:pricebookId/versions/:versionId/items", items.Upsert)

	return rg
}
