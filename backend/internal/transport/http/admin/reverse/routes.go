package reverse

import (
	reversesvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/reverse"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts reverse logistics admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/reverse", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(reversesvc.NewWaybillService(deps))
	rg.GET("/waybills", handler.ListWaybills)
	rg.POST("/waybills", handler.CreateWaybill)
	rg.GET("/waybills/:id", handler.GetWaybill)
	rg.POST("/waybills/:id/track", handler.AppendTracking)
	rg.POST("/waybills/:id/warehouse-result", handler.RecordWarehouseResult)
	return rg
}
