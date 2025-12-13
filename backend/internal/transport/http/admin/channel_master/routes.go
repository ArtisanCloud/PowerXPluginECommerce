package channel_master

import (
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel_master"
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
	if deps != nil && deps.DB != nil {
		audit := channelobs.NewAuditEmitter(deps.RuntimeLogger(nil, "channel-master-audit", nil))
		service := channelservice.NewService(deps, nil, audit)
		handler = NewHandler(service)
	} else {
		handler = NewHandler(nil)
	}
	channelsGroup.GET("", handler.List)
	channelsGroup.POST("", handler.Create)
}
