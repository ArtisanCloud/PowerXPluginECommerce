package customer

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries exposes capability mapping for customer APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/customers"
	return map[string]authx.Permission{
		"GET:" + base:                      {Resource: "customer.read", Action: "read"},
		"GET:" + base + "/import/template": {Resource: "customer.read", Action: "read"},
		"GET:" + base + "/import/conflicts/:taskId": {Resource: "customer.read", Action: "read"},
		"GET:" + base + "/:id/entitlements": {Resource: "customer.read", Action: "read"},
		"GET:" + base + "/:id/tokens":       {Resource: "customer.read", Action: "read"},
		"POST:" + base:                     {Resource: "customer.manage", Action: "write"},
		"POST:" + base + "/import":         {Resource: "customer.manage", Action: "write"},
		"PATCH:" + base + "/:id":           {Resource: "customer.manage", Action: "write"},
		"DELETE:" + base + "/:id":          {Resource: "customer.delete", Action: "delete"},
	}
}
