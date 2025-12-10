package customer

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes attaches customer domain APIs.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil {
		return
	}
	handler := NewHandler(deps)
	group := rg.Group("/customers", httpmw.EnsureTenant())
	{
		group.GET("", handler.ListCustomers)
		group.POST("", handler.CreateCustomer)
		group.PATCH(":id", handler.UpdateCustomer)
		group.DELETE(":id", handler.DeleteCustomer)
	}
}
