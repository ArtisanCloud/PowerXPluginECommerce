package security

import (
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires the admin security namespace.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil || deps == nil || deps.Config == nil {
		return
	}
	auditWriter := CreateAuditWriter(deps.Config)
	consent := NewConsentHandler(deps, auditWriter)
	toolgrant := NewToolGrantHandler(deps)
	audit := NewAuditReportHandler(deps, auditWriter)
	advisory := NewAdvisoryHandler(deps)
	sec := rg.Group("/security")
	{
		sec.GET("/consent-tokens", consent.ListConsentTokens)
		sec.POST("/consent-tokens/:tokenId/revoke", consent.RevokeConsentToken)
		sec.GET("/lifecycle-events", consent.ListLifecycleEvents)
		sec.GET("/audit-reports", audit.ListReports)
		sec.POST("/advisories", advisory.Create)
		sec.GET("/advisories", advisory.List)
		sec.POST("/advisories/:advisoryId/publish", advisory.Publish)
		sec.POST("/toolgrants/revoke", toolgrant.Revoke)
		sec.GET("/toolgrants/revocations", toolgrant.ListRevocations)
		sec.GET("/toolgrants/usage", toolgrant.ListUsageEvents)
	}
}

// RBACEntries returns RBAC metadata for admin security endpoints.
func RBACEntries(prefix string) map[string]authx.Permission {
	if prefix == "" {
		prefix = "/api/v1"
	}
	base := prefix + "/admin/security"
	return map[string]authx.Permission{
		"GET:" + base + "/consent-tokens":                  {Resource: "admin.security.consent", Action: "read"},
		"POST:" + base + "/consent-tokens/:tokenId/revoke": {Resource: "admin.security.consent", Action: "write"},
		"GET:" + base + "/lifecycle-events":                {Resource: "admin.security.lifecycle", Action: "read"},
		"GET:" + base + "/audit-reports":                   {Resource: "admin.security.audit", Action: "read"},
		"GET:" + base + "/advisories":                      {Resource: "admin.security.advisory", Action: "read"},
		"POST:" + base + "/advisories":                     {Resource: "admin.security.advisory", Action: "write"},
		"POST:" + base + "/advisories/:advisoryId/publish": {Resource: "admin.security.advisory", Action: "write"},
		"POST:" + base + "/toolgrants/revoke":              {Resource: "admin.security.toolgrant", Action: "write"},
		"GET:" + base + "/toolgrants/revocations":          {Resource: "admin.security.toolgrant", Action: "read"},
		"GET:" + base + "/toolgrants/usage":                {Resource: "admin.security.toolgrant", Action: "read"},
	}
}
