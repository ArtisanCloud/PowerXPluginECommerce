package product_category

import (
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/product/categories and related endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}
	productGroup := router.Group("/product", httpmw.EnsureTenant())

	var handler *Handler
	if deps != nil && deps.DB != nil {
		handler = NewHandler(categoryservice.NewService(deps))
	} else {
		handler = NewHandler(nil)
	}

	// Category CRUD + tree management.
	categories := productGroup.Group("/categories")
	categories.GET("/tree", handler.Tree)
	categories.GET("", handler.List)
	categories.POST("", handler.Create)
	categories.PATCH("/:id", handler.Update)
	categories.DELETE("/:id", handler.Delete)
	categories.POST("/:id/move", handler.Move)
	categories.PATCH("/:id/status", handler.SetStatus)
	categories.POST("/import", handler.CategoriesImport)
	categories.POST("/export", handler.CategoriesExport)
	categories.GET("/:id/mappings", handler.MappingList)
	categories.POST("/:id/mappings", handler.MappingUpsert)
	categories.GET("/:id/audit", handler.CategoryAudit)

	templates := productGroup.Group("/category-templates")
	templates.GET("", handler.TemplateList)
	templates.POST("", handler.TemplateCreate)
	templates.GET("/effective", handler.TemplateEffective)
	templates.GET("/:id", handler.TemplateGet)
	templates.PATCH("/:id", handler.TemplateUpdate)
	templates.POST("/:id/publish", handler.TemplatePublish)
	templates.POST("/:id/rollback", handler.TemplateRollback)
	templates.GET("/:id/versions", handler.TemplateVersions)
	templates.GET("/:id/preview", handler.TemplatePreview)
	templates.POST("/:id/simulate", handler.TemplateSimulate)
	templates.GET("/:id/impact", handler.TemplateImpact)
	templates.POST("/:id/impact/recheck", handler.TemplateTriggerRecheck)
}
