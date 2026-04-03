package coupon

import (
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/coupons namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/coupons", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	templateHandler := NewTemplateHandler(couponsvc.NewTemplateService(deps))
	issueHandler := NewIssueHandler(couponsvc.NewIssueService(deps))
	queryHandler := NewQueryHandler(couponsvc.NewQueryService(deps))

	rg.GET("/templates", templateHandler.List)
	rg.POST("/templates", templateHandler.Create)
	rg.PATCH("/templates/:id", templateHandler.Update)

	rg.POST("/issues", issueHandler.Issue)

	rg.GET("/assets", queryHandler.ListAssets)
	rg.GET("/usage-logs", queryHandler.ListUsageLogs)
	return rg
}
