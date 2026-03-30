package membership

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares admin membership route permissions.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/membership"
	res := func() string { return "com.powerx.plugins.ecommerce:membership" }
	return map[string]authx.Permission{
		"GET:" + base + "/tiers":                {Resource: res(), Action: "read"},
		"POST:" + base + "/tiers":               {Resource: res(), Action: "write"},
		"PATCH:" + base + "/tiers/:id/status":   {Resource: res(), Action: "write"},
		"DELETE:" + base + "/tiers/:id":         {Resource: res(), Action: "write"},
		"GET:" + base + "/benefits":             {Resource: res(), Action: "read"},
		"POST:" + base + "/benefits":            {Resource: res(), Action: "write"},
		"POST:" + base + "/entitlements/grant":  {Resource: res(), Action: "write"},
		"POST:" + base + "/entitlements/revoke": {Resource: res(), Action: "write"},
		"POST:" + base + "/tokens/adjust":       {Resource: res(), Action: "write"},
		"GET:" + base + "/tokens/transactions":  {Resource: res(), Action: "read"},
		"POST:" + base + "/points/redeem":       {Resource: res(), Action: "write"},
	}
}
