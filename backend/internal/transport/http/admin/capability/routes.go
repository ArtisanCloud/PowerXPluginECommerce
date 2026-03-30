package capability

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires capability endpoints under /admin.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil {
		return
	}
	handler := NewHandler(deps)

	rg.GET("/capabilities", handler.ListCatalog)
	rg.GET("/capabilities/sources", handler.ListSources)

	register := rg.Group("/capabilities/register")
	{
		register.GET("/template", handler.GetRegisterTemplate)
		register.POST("/validate", handler.ValidateDraft)
		register.POST("", handler.Submit)
	}

	exposure := rg.Group("/capabilities/exposure")
	{
		exposure.GET("/template", handler.GetExposureTemplate)
		exposure.GET("/:capabilityID", handler.GetExposurePackage)
		exposure.PUT("/:capabilityID", handler.UpsertExposurePackage)
	}

	quotas := rg.Group("/capabilities/quotas")
	{
		quotas.GET("/:capabilityID", handler.ListQuotas)
		quotas.POST("/:capabilityID", handler.UpsertQuota)
	}

	lifecycle := rg.Group("/capabilities/lifecycle")
	{
		lifecycle.GET("/template", handler.GetLifecycleTemplate)
		lifecycle.GET("", handler.ListLifecyclePlans)
		lifecycle.POST("", handler.CreateLifecyclePlan)
		lifecycle.POST("/:planID/status", handler.UpdateLifecycleStatus)
	}
}
