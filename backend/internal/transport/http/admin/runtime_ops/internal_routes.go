package runtime_ops

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterInternalRoutes wires host-style internal routes under /api/v1/internal.
func RegisterInternalRoutes(router *gin.RouterGroup, deps *app.Deps) {
	if router == nil {
		return
	}
	wsBus := NewWSBusHandler(deps)
	wsg := router.Group("/ws-bus")
	wsg.POST("/grant", wsBus.Grant)
	wsg.POST("/publish", wsBus.Publish)

	eventFabric := NewEventFabricHandler(deps)
	efg := router.Group("/event-fabric")
	efg.POST("/topics", eventFabric.CreateTopics)
}
