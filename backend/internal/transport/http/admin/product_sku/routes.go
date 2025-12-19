package product_sku

import (
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes attaches SKU endpoints beneath /admin/product.
func RegisterRoutes(productGroup *gin.RouterGroup, deps *app.Deps) {
	if productGroup == nil {
		return
	}

	var service *productskuservice.Service
	if deps != nil {
		service = productskuservice.NewService(deps)
	}

	skuHandler := NewHandler(service)
	bulkHandler := NewBulkHandler(service)
	channelHandler := NewChannelHandler(service)
	importHandler := NewImportExportHandler(service)
	inventoryHandler := NewInventoryHandler(service)
	serialHandler := NewSerialHandler(service)
	generatorHandler := NewGeneratorHandler(service)

	skus := productGroup.Group("/skus", httpmw.EnsureTenant())
	{
		skus.GET("", skuHandler.List)
		skus.POST("", skuHandler.Upsert)
		skus.PATCH("/:id", skuHandler.Update)
		skus.DELETE("/:id", skuHandler.Delete)
		skus.POST("/bulk-tasks", bulkHandler.Submit)
		skus.GET("/bulk-tasks/:taskId", bulkHandler.Get)
		skus.POST("/import", importHandler.Import)
		skus.POST("/export", importHandler.Export)
		skus.GET("/:id/inventory", inventoryHandler.Get)
		skus.GET("/:id/channels", channelHandler.List)
		skus.POST("/:id/channels", channelHandler.Upsert)
		skus.POST("/:id/channels/publish", channelHandler.Publish)
		skus.GET("/:id/serials", serialHandler.List)
		skus.POST("/:id/serials", serialHandler.Create)
	}

	// SKU generator endpoint lives under the SPU scope per contract.
	productGroup.POST("/spus/:id/skus/generate", httpmw.EnsureTenant(), generatorHandler.Generate)
}
