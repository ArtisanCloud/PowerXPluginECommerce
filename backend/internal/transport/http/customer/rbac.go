package customer

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries exposes capability mapping for customer APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/customers"
	return map[string]authx.Permission{
		"GET:" + base: {Resource: "customer.list", Action: "read"},
	}
}
