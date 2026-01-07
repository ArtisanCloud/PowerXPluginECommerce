package product_spec

import (
	specservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_spec"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/product/spus/:id/spec-groups endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}

	var handler *Handler
	if deps != nil && deps.DB != nil {
		handler = NewHandler(specservice.NewService(deps))
	} else {
		handler = NewHandler(nil)
	}

	spus := router.Group("/spus")
	{
		spus.GET("/:id/spec-groups", handler.List)
		spus.PUT("/:id/spec-groups", handler.Replace)
	}
}

