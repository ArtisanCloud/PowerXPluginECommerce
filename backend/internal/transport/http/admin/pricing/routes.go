package pricing

import (
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/pricing namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/pricing")
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

	rg.POST("/pricebooks/:pricebookId/versions", versions.Create)
	rg.POST("/pricebooks/:pricebookId/versions/:versionId/publish", versions.Publish)
	rg.POST("/pricebooks/:pricebookId/versions/:versionId/archive", versions.Archive)
	rg.PUT("/pricebooks/:pricebookId/versions/:versionId/items", items.Upsert)

	return rg
}
