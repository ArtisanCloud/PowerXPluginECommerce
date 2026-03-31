package after_sales

import (
	adminsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts after-sales admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/after-sales", httpmw.EnsureTenant())
	if deps == nil || deps.DB == nil {
		return rg
	}

	caseService := adminsvc.NewCaseService(deps)
	decisionService := adminsvc.NewDecisionService(deps)
	dashboardService := adminsvc.NewDashboardService(deps)
	reverseLinkService := adminsvc.NewReverseLinkService(deps)

	handler := NewHandler(caseService, decisionService)
	dashboard := NewDashboardHandler(dashboardService)
	reverse := NewReverseLogisticsHandler(reverseLinkService)

	rg.GET("/cases", handler.ListCases)
	rg.GET("/cases/:id", handler.GetCase)
	rg.POST("/cases/:id/accept", handler.AcceptCase)
	rg.POST("/cases/:id/review", handler.ReviewCase)
	rg.POST("/cases/:id/approve", handler.ApproveCase)
	rg.POST("/cases/:id/reject", handler.RejectCase)
	rg.POST("/cases/:id/complete", handler.CompleteCase)
	rg.POST("/cases/:id/close", handler.CloseCase)
	rg.POST("/cases/:id/reverse-logistics", reverse.LinkCase)
	rg.GET("/dashboard", dashboard.GetSnapshot)

	return rg
}
