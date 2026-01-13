package miniapp

import (
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	miniappauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/auth"
	miniappcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/category"
	miniapporder "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/order"
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
	// Open (no customer token required): 用于小程序“游客态”浏览类目/商品列表。
	open := group.Group("")
	miniappcategory.RegisterRoutes(open, deps)
	product.RegisterRoutes(open, deps)

	// Protected (customer token required): 预留给后续订单、收藏等需要用户态的能力。
	protected := group.Group("")
	protected.Use(httpmw.CustomerAuthenticate(authenticator))
	miniapporder.RegisterRoutes(protected, deps)
	return group
}
