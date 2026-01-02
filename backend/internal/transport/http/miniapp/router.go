package miniapp

import (
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	miniappauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/auth"
	miniappcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/product"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires the /mini-app prefix used by mobile/mini-program entrypoints.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if rg == nil {
		return nil
	}
	var authenticator customerauth.Authenticator
	if deps != nil {
		authenticator = deps.CustomerAuthenticator
	}
	group := rg.Group("/mini-app", httpmw.EnsureTenant())
	miniappauth.RegisterRoutes(group, deps)
	protected := group.Group("")
	protected.Use(httpmw.CustomerAuthenticate(authenticator))
	miniappcategory.RegisterRoutes(protected, deps)
	product.RegisterRoutes(protected, deps)
	return group
}
