package admin

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	adminchannels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/channel_master"
	adminconsole "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/console"
	adminintegration "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/integration"
	adminmarketplace "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/marketplace"
	adminoperations "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/operations"
	adminpricing "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/pricing"
	adminproduct "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product"
	adminproductcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/product_category"
	adminruntime "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/runtime_ops"
	adminsecurity "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/security"
	"github.com/gin-gonic/gin"
)

// Register 注册 Admin 路由
func RegisterAPIRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	adminHandler := NewAdminHandler(deps)
	admin := rg.Group("/admin")
	{
		// 基础管理功能
		admin.GET("/manifest", adminHandler.GetManifest) // 获取插件清单
		admin.GET("/rbac", adminHandler.GetRBACInfo)     // 获取权限信息

		runtimeOps := admin.Group("/runtime")
		adminruntime.RegisterRoutes(runtimeOps, deps)

		adminmarketplace.RegisterRoutes(admin, deps)
		adminoperations.RegisterRoutes(admin, deps)
		adminproduct.RegisterRoutes(admin, deps)
		adminproductcategory.RegisterRoutes(admin, deps)
		adminconsole.RegisterRoutes(admin, deps)
		adminintegration.RegisterRoutes(admin, deps)
		adminsecurity.RegisterRoutes(admin, deps)
		adminchannels.RegisterRoutes(admin, deps)
		adminpricing.RegisterRoutes(admin, deps)
	}
}
