package admin

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	adminaftersales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/after_sales"
	admincapability "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/capability"
	adminchannels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/channel_master"
	adminconsole "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/console"
	admincustomeraddress "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/customer_address"
	adminfulfillment "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/fulfillment"
	adminintegration "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/integration"
	adminlogistics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/logistics"
	adminmarketplace "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/marketplace"
	adminmembership "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/membership"
	adminoperations "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/operations"
	adminorder "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/order"
	adminpayments "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/payments"
	adminpricing "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/pricing"
	adminproduct "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product"
	adminproductcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product_category"
	adminreverse "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/reverse"
	adminruntime "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/runtime_ops"
	adminsecurity "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/security"
	adminsubscriptionreconciliation "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/subscription_reconciliation"
	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers admin routes.
func RegisterAPIRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	adminHandler := NewAdminHandler(deps)
	admin := rg.Group("/admin")
	{
		admin.GET("/manifest", adminHandler.GetManifest)
		admin.GET("/rbac", adminHandler.GetRBACInfo)

		runtimeOps := admin.Group("/runtime")
		adminruntime.RegisterRoutes(runtimeOps, deps)

		internal := rg.Group("/internal")
		adminruntime.RegisterInternalRoutes(internal, deps)

		adminmarketplace.RegisterRoutes(admin, deps)
		adminoperations.RegisterRoutes(admin, deps)
		adminproduct.RegisterRoutes(admin, deps)
		adminproductcategory.RegisterRoutes(admin, deps)
		adminconsole.RegisterRoutes(admin, deps)
		adminintegration.RegisterRoutes(admin, deps)
		adminsecurity.RegisterRoutes(admin, deps)
		adminchannels.RegisterRoutes(admin, deps)
		admincapability.RegisterRoutes(admin, deps)
		adminpricing.RegisterRoutes(admin, deps)
		adminorder.RegisterRoutes(admin, deps)
		adminpayments.RegisterRoutes(admin, deps)
		admincustomeraddress.RegisterRoutes(admin, deps)
		adminmembership.RegisterRoutes(admin, deps)
		adminaftersales.RegisterRoutes(admin, deps)
		adminlogistics.RegisterRoutes(admin, deps)
		adminfulfillment.RegisterRoutes(admin, deps)
		adminreverse.RegisterRoutes(admin, deps)
		adminsubscriptionreconciliation.RegisterRoutes(admin, deps)
	}
}
