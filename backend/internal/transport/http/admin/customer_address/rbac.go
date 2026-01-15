package customer_address

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for customer address book management.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/customers/*/addresses"
	return map[string]authx.Permission{
		"GET:" + base:                 {Resource: "customer.addresses", Action: "read"},
		"POST:" + base:                {Resource: "customer.addresses", Action: "write"},
		"PATCH:" + base + "/*":        {Resource: "customer.addresses", Action: "write"},
		"DELETE:" + base + "/*":       {Resource: "customer.addresses", Action: "write"},
		"POST:" + base + "/*/default": {Resource: "customer.addresses", Action: "write"},
	}
}
