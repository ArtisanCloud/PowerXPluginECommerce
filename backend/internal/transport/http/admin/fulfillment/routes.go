package fulfillment

import (
	fulfillmentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/fulfillment"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts fulfillment admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/fulfillment", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(
		fulfillmentsvc.NewTaskService(deps),
		fulfillmentsvc.NewWaveService(deps),
		fulfillmentsvc.NewExceptionService(deps),
	)
	rg.GET("/tasks", handler.ListTasks)
	rg.POST("/tasks", handler.CreateTask)
	rg.PATCH("/tasks/:id/complete", handler.CompleteTask)
	rg.GET("/waves", handler.ListWaves)
	rg.POST("/waves", handler.CreateWave)
	rg.GET("/waves/:id", handler.GetWave)
	rg.PATCH("/waves/:id/advance", handler.AdvanceWave)
	rg.PATCH("/waves/:id/tasks/:task_id/reassign", handler.ReassignWaveTask)
	rg.GET("/exceptions", handler.ListExceptions)
	rg.POST("/exceptions", handler.ReportException)
	return rg
}
