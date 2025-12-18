package channel_master

import (
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /admin/channels endpoints.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}
	channelsGroup := router.Group("/channels", httpmw.EnsureTenant())
	var handler *Handler
	var credentialHandler *CredentialHandler
	var detailHandler *DetailHandler
	var alertsHandler *AlertsHandler
	var syncHandler *SyncHandler
	var taskNoteHandler *TaskNoteHandler
	var strategyHandler *StrategyHandler
	var metricsCollector *channelobs.Metrics
	if deps != nil && deps.DB != nil {
		audit := channelobs.NewAuditEmitter(deps.RuntimeLogger(nil, "channel-master-audit", nil))
		metricsCollector = channelobs.NewMetrics(deps.RuntimeLogger(nil, "channel-master-metrics", nil))
		service := channelservice.NewService(deps, nil, audit, metricsCollector)
		ownerDirectory := channelservice.NewOwnerDirectory(deps)
		handler = NewHandler(service, ownerDirectory)
		alertEmitter := channelobs.NewAlertEmitter(deps.RuntimeLogger(nil, "channel-master-alert", nil))
		credSvc := channelservice.NewCredentialService(deps, alertEmitter, metricsCollector)
		credentialHandler = NewCredentialHandler(credSvc)
		metricsSvc := channelservice.NewMetricsService(deps, metricsCollector)
		taskSvc := channelservice.NewTaskNoteService(deps, audit)
		syncSvc := channelservice.NewSyncHistoryService(deps, metricsCollector)
		strategySvc := channelservice.NewStrategyService(deps, audit)
		detailHandler = NewDetailHandler(deps, metricsSvc, taskSvc, syncSvc, credSvc, strategySvc)
		alertsHandler = NewAlertsHandler(deps)
		syncHandler = NewSyncHandler(syncSvc)
		taskNoteHandler = NewTaskNoteHandler(taskSvc)
		strategyHandler = NewStrategyHandler(strategySvc)
	} else {
		handler = NewHandler(nil, nil)
		credentialHandler = NewCredentialHandler(nil)
		detailHandler = &DetailHandler{}
		alertsHandler = &AlertsHandler{}
		syncHandler = &SyncHandler{}
		taskNoteHandler = &TaskNoteHandler{}
		strategyHandler = NewStrategyHandler(nil)
	}
	channelsGroup.GET("", handler.List)
	channelsGroup.GET("/platforms", handler.Platforms)
	channelsGroup.GET("/owners", handler.Owners)
	channelsGroup.POST("", handler.Create)
	channelsGroup.PATCH("/:channelId", handler.Update)
	channelsGroup.POST("/:channelId/submit", handler.Submit)
	channelsGroup.POST("/:channelId/approval", handler.Approve)
	channelsGroup.GET("/:channelId", detailHandler.Get)
	channelsGroup.GET("/:channelId/credentials", credentialHandler.List)
	channelsGroup.POST("/:channelId/credentials", credentialHandler.Upsert)
	channelsGroup.POST("/:channelId/credentials/test", credentialHandler.Test)
	channelsGroup.GET("/:channelId/alerts", alertsHandler.List)
	channelsGroup.PATCH("/:channelId/alerts/:alertId", alertsHandler.Update)
	channelsGroup.POST("/:channelId/sync", syncHandler.Trigger)
	channelsGroup.GET("/:channelId/sync-history", syncHandler.History)
	channelsGroup.GET("/:channelId/tasks", taskNoteHandler.ListTasks)
	channelsGroup.POST("/:channelId/tasks", taskNoteHandler.LinkTask)
	channelsGroup.PATCH("/:channelId/tasks/:taskLinkId", taskNoteHandler.UpdateTask)
	channelsGroup.DELETE("/:channelId/tasks/:taskLinkId", taskNoteHandler.RemoveTask)
	channelsGroup.GET("/:channelId/notes", taskNoteHandler.ListNotes)
	channelsGroup.POST("/:channelId/notes", taskNoteHandler.AddNote)
	channelsGroup.GET("/:channelId/strategy", strategyHandler.Get)
	channelsGroup.PATCH("/:channelId/strategy", strategyHandler.Upsert)
}
