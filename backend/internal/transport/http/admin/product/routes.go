package product

import (
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	productspec "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product_spec"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product/spu"
	productsku "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product_sku"
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
		handler             *spu.Handler
		skuHandler          *spu.SKUHandler
		versionHandler      *spu.VersionHandler
		channelHandler      *spu.ChannelHandler
		planHandler         *spu.PlanHandler
		importExportHandler *spu.ImportExportHandler
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
		planService := spuservice.NewSubscriptionPlanService(deps)
		planHandler = spu.NewPlanHandler(planService)
		importService := spuservice.NewImportService(deps)
		importExportHandler = spu.NewImportExportHandler(importService)
	} else {
		handler = spu.NewHandler(nil)
		skuHandler = spu.NewSKUHandler(nil)
		versionHandler = spu.NewVersionHandler(nil)
		channelHandler = spu.NewChannelHandler(nil)
		planHandler = spu.NewPlanHandler(nil)
		importExportHandler = spu.NewImportExportHandler(nil)
	}

	spus := productGroup.Group("/spus")
	{
		spus.GET("", handler.List)
		spus.POST("", handler.Create)
		spus.PATCH("/:id", handler.Update)
		spus.GET("/:id", handler.Get)
		spus.POST("/:id/revise", handler.Revise)
		spus.POST("/:id/submit", handler.Submit)
		spus.POST("/:id/publish", handler.Publish)
		spus.POST("/:id/withdraw", handler.Withdraw)
		spus.POST("/:id/delete", handler.Delete)
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
		spus.GET("/:id/subscription-plans", planHandler.List)
		spus.POST("/:id/subscription-plans", planHandler.Create)
		spus.PATCH("/:id/subscription-plans/:planId", planHandler.Update)
		spus.DELETE("/:id/subscription-plans/:planId", planHandler.Delete)
		spus.POST("/import", importExportHandler.Import)
		spus.POST("/export", importExportHandler.Export)
	}

	productspec.RegisterRoutes(productGroup, deps)

	// SKU routes are registered independently to keep handler boundaries clear.
	productsku.RegisterRoutes(productGroup, deps)
}
