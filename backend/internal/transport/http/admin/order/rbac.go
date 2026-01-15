package order

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for order admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/orders"
	res := func() string { return "com.powerx.plugin.ecommerce:order" }
	return map[string]authx.Permission{
		"GET:" + base:                  {Resource: res(), Action: "read"},
		"POST:" + base:                 {Resource: res(), Action: "create"},
		"GET:" + base + "/:id":         {Resource: res(), Action: "read"},
		"POST:" + base + "/:id/cancel": {Resource: res(), Action: "cancel"},
	}
}
