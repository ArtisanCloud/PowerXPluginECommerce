package promotion

import (
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/promotions namespace.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/promotions", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	campaigns := NewCampaignHandler(promotionsvc.NewCampaignService(deps))
	audits := NewAuditHandler(promotionsvc.NewAuditLogService(deps))
	rg.GET("", campaigns.List)
	rg.POST("", campaigns.Create)
	rg.PATCH("/:id", campaigns.Update)
	rg.POST("/:id/activate", campaigns.Activate)
	rg.POST("/:id/pause", campaigns.Pause)
	rg.POST("/:id/clone", campaigns.Clone)
	rg.GET("/:id/audit-logs", audits.List)
	return rg
}
