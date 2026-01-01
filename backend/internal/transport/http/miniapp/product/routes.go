package product

import (
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	skuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes exposes /mini-app/products endpoints via the product domain module.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}

	var handler *Handler
	if deps != nil && deps.DB != nil {
		handler = NewHandler(
			spuservice.NewService(deps),
			skuservice.NewService(deps),
		)
	} else {
		handler = NewHandler(nil, nil)
	}

	group := router.Group("/products")
	group.GET("", handler.ListProducts)
	group.GET("/:id/skus", handler.ListSkus)

	return group
}
