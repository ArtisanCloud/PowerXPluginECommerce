package category

import (
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes exposes /mini-app/categories endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	var handler *Handler
	if deps != nil && deps.DB != nil {
		handler = NewHandler(categoryservice.NewService(deps))
	} else {
		handler = NewHandler(nil)
	}
	group := router.Group("/categories")
	group.GET("/tree", handler.Tree)
	return group
}
