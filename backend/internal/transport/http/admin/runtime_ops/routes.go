package runtime_ops

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	runtimeops "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/runtime_ops"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires runtime ops endpoints behind the admin router.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil || deps == nil || deps.Config == nil {
		return
	}
	bootstrap := NewBootstrapHandler(runtimeops.NewService())
	router.POST("/bootstrap", bootstrap.Bootstrap)
	loggingPolicy := NewLoggingPolicyHandler()
	router.GET("/logging/policy", loggingPolicy.Get)
	router.PUT("/logging/policy", loggingPolicy.Put)
	router.POST("/logging/probe", LoggingProbeHandler())

	sessions := NewSessionsHandler(deps)
	router.POST("/sessions/register", sessions.Register)

	runtimeOps := deps.Config.RuntimeOps
	if runtimeOps == nil {
		runtimeOps = &config.RuntimeOpsDefaults{}
	}
	quotaHandler := NewQuotaHandler(deps, runtimeOps)
	quota := router.Group("/quota")
	quota.GET("/status", quotaHandler.GetStatus)
	quota.POST("/overrides", quotaHandler.SetOverride)

	wsBus := NewWSBusHandler(deps)
	internal := router.Group("/internal/ws-bus")
	internal.POST("/grant", wsBus.Grant)
	internal.POST("/publish", wsBus.Publish)

	eventFabric := NewEventFabricHandler(deps)
	internalEventFabric := router.Group("/internal/event-fabric")
	internalEventFabric.POST("/topics", eventFabric.CreateTopics)

	router.GET("/metrics", MetricsHandler)
}
