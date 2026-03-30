package order

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for order admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/orders"
	res := func() string { return "com.powerx.plugins.ecommerce:order" }
	return map[string]authx.Permission{
		"GET:" + base:                             {Resource: res(), Action: "read"},
		"POST:" + base:                            {Resource: res(), Action: "create"},
		"GET:" + base + "/:id":                    {Resource: res(), Action: "read"},
		"POST:" + base + "/:id/cancel":            {Resource: res(), Action: "cancel"},
		"PATCH:" + base + "/:id/shipping-address": {Resource: res(), Action: "update"},
		"GET:" + base + "/:id/benefit-reviews":    {Resource: res(), Action: "read"},
		"POST:" + base + "/:id/benefit-reviews":   {Resource: res(), Action: "update"},
		"POST:" + base + "/benefit-reviews/approve": {Resource: res(), Action: "update"},
		"POST:" + base + "/benefit-reviews/reject":  {Resource: res(), Action: "update"},
		"GET:" + base + "/benefit-codes":          {Resource: res(), Action: "read"},
	}
}
