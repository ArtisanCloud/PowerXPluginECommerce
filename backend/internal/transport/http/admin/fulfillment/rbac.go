package fulfillment

import (
	"strings"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for fulfillment admin APIs.
func RBACEntries(prefix string) map[string]AuthX.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/fulfillment"
	res := func(suffix string) string {
		return "com.powerx.plugins.ecommerce:" + suffix
	}
	return map[string]AuthX.Permission{
		"GET:" + base + "/tasks":                {Resource: res("fulfillment.task"), Action: "read"},
		"POST:" + base + "/tasks":               {Resource: res("fulfillment.task"), Action: "manage"},
		"PATCH:" + base + "/tasks/:id/complete": {Resource: res("fulfillment.task"), Action: "manage"},
		"GET:" + base + "/exceptions":           {Resource: res("fulfillment.exception"), Action: "read"},
		"POST:" + base + "/exceptions":          {Resource: res("fulfillment.exception"), Action: "manage"},
	}
}
