package product

import (
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/product/* endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}
	productGroup := router.Group("/product")

	var handler *spu.Handler
	if deps != nil && deps.DB != nil {
		service := spuservice.NewService(deps)
		handler = spu.NewHandler(service)
	} else {
		handler = spu.NewHandler(nil)
	}

	spus := productGroup.Group("/spus")
	{
		spus.GET("", handler.List)
		spus.POST("", handler.Create)
		spus.GET(":id", handler.Get)
	}
}
