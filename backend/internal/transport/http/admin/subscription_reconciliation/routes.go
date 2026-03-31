package subscription_reconciliation

import (
	SubscriptionReconciliationSvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/subscription_reconciliation"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts subscription reconciliation admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/subscription-reconciliation", httpmw.EnsureTenant())
	service := SubscriptionReconciliationSvc.NewService(deps)
	handler := NewHandler(service)

	rg.POST("/batches", handler.CreateBatch)
	rg.GET("/batches", handler.ListBatches)
	rg.GET("/batches/:id/deltas", handler.ListBatchDeltas)
	rg.POST("/deltas/:id/tasks", handler.CreateDeltaTask)
	rg.POST("/tasks/:id/close", handler.CloseTask)
	rg.POST("/governance/run", handler.RunGovernance)
	rg.GET("/dashboard", handler.Dashboard)

	return rg
}
