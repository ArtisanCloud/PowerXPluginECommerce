package product

import (
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product/spu"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/product/* endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}
	productGroup := router.Group("/product", httpmw.EnsureTenant())

	var (
		handler        *spu.Handler
		skuHandler     *spu.SKUHandler
		versionHandler *spu.VersionHandler
		channelHandler *spu.ChannelHandler
	)
	if deps != nil && deps.DB != nil {
		service := spuservice.NewService(deps)
		handler = spu.NewHandler(service)
		skuService := spuservice.NewSKULinkService(deps)
		skuHandler = spu.NewSKUHandler(skuService)
		versionService := spuservice.NewVersionService(deps)
		versionHandler = spu.NewVersionHandler(versionService)
		channelService := spuservice.NewChannelService(deps)
		channelHandler = spu.NewChannelHandler(channelService)
	} else {
		handler = spu.NewHandler(nil)
		skuHandler = spu.NewSKUHandler(nil)
		versionHandler = spu.NewVersionHandler(nil)
		channelHandler = spu.NewChannelHandler(nil)
	}

	spus := productGroup.Group("/spus")
	{
		spus.GET("", handler.List)
		spus.POST("", handler.Create)
		spus.PATCH("/:id", handler.Update)
		spus.GET("/:id", handler.Get)
		spus.POST("/:id/submit", handler.Submit)
		spus.POST("/:id/publish", handler.Publish)
		spus.GET("/:id/skus", skuHandler.List)
		spus.PUT("/:id/skus", skuHandler.Replace)
		spus.GET("/:id/versions", versionHandler.List)
		spus.GET("/:id/versions/:versionId", versionHandler.Get)
		spus.POST("/:id/versions/:versionId/approve", versionHandler.Approve)
		spus.POST("/:id/versions/:versionId/reject", versionHandler.Reject)
		spus.POST("/:id/versions/:versionId/rollback", versionHandler.Rollback)
		spus.GET("/:id/channels", channelHandler.List)
		spus.POST("/:id/channels", channelHandler.Upsert)
		spus.DELETE("/:id/channels/:channel", channelHandler.Delete)
	}
}
