package templates

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for template CRUD APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/templates"
	return map[string]authx.Permission{
		"GET:" + base:             {Resource: "com.powerx.plugins.ecommerce:template", Action: "read"},
		"GET:" + base + "/:id":    {Resource: "com.powerx.plugins.ecommerce:template", Action: "read"},
		"POST:" + base:            {Resource: "com.powerx.plugins.ecommerce:template", Action: "manage"},
		"PUT:" + base + "/:id":    {Resource: "com.powerx.plugins.ecommerce:template", Action: "manage"},
		"DELETE:" + base + "/:id": {Resource: "com.powerx.plugins.ecommerce:template", Action: "manage"},
	}
}
