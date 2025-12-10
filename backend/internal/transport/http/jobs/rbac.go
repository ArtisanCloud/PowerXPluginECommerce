package jobs

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries exposes permissions for /jobs endpoints.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/jobs"
	return map[string]authx.Permission{
		"GET:" + base + "/*": {Resource: "customer.read", Action: "read"},
	}
}
