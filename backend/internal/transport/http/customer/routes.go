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
		group.GET("import/template", handler.DownloadImportTemplate)
		group.POST("import", handler.CreateImportTask)
		group.POST("", handler.CreateCustomer)
		group.GET(":id", handler.GetCustomer)
		group.GET(":id/entitlements", handler.ListCustomerEntitlements)
		group.GET(":id/tokens", handler.GetCustomerTokenBalances)
		group.PATCH(":id", handler.UpdateCustomer)
		group.DELETE(":id", handler.DeleteCustomer)
	}
}
