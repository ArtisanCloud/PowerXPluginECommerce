package membership

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares admin membership route permissions.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/membership"
	res := func() string { return "com.powerx.plugin.ecommerce:membership" }
	return map[string]authx.Permission{
		"GET:" + base + "/tiers":              {Resource: res(), Action: "read"},
		"GET:" + base + "/benefits":           {Resource: res(), Action: "read"},
		"POST:" + base + "/entitlements/grant": {Resource: res(), Action: "write"},
		"POST:" + base + "/entitlements/revoke": {Resource: res(), Action: "write"},
		"POST:" + base + "/tokens/adjust":     {Resource: res(), Action: "write"},
	}
}
