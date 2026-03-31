package miniapp

import (
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	miniappaftersales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/after_sales"
	miniappauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/auth"
	miniappcart "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/cart"
	miniappcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/category"
	miniappaddress "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/customer_address"
	miniappmembership "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/membership"
	miniapporder "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/order"
	miniapppayments "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp/payments"
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
	group := rg.Group("/mini-app")
	miniappauth.RegisterRoutes(group, deps)
	// Open (no customer token required): 用于小程序“游客态”浏览类目/商品列表。
	open := group.Group("", httpmw.EnsureTenant())
	miniappcategory.RegisterRoutes(open, deps)
	product.RegisterRoutes(open, deps)
	miniapppayments.RegisterRoutes(open, nil, deps)

	// Protected (customer token required): 预留给后续订单、收藏等需要用户态的能力。
	protected := group.Group("")
	protected.Use(httpmw.CustomerAuthenticate(authenticator))
	protected.Use(httpmw.EnsureTenant())
	miniappcart.RegisterRoutes(protected, deps)
	miniapporder.RegisterRoutes(protected, deps)
	miniappaddress.RegisterRoutes(protected, deps)
	miniappmembership.RegisterRoutes(protected, deps)
	miniappaftersales.RegisterRoutes(protected, deps)
	miniapppayments.RegisterRoutes(nil, protected, deps)
	return group
}
