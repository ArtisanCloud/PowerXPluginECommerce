package capability

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries returns capability admin route permission mappings.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/capabilities"
	resource := "capability.management"
	return map[string]authx.Permission{
		"GET:" + base:                                {Resource: resource, Action: "read"},
		"GET:" + base + "/sources":                   {Resource: resource, Action: "read"},
		"GET:" + base + "/register/template":         {Resource: resource, Action: "read"},
		"POST:" + base + "/register/validate":        {Resource: resource, Action: "write"},
		"POST:" + base + "/register":                 {Resource: resource, Action: "write"},
		"GET:" + base + "/exposure/template":         {Resource: resource, Action: "read"},
		"GET:" + base + "/exposure/:capabilityID":    {Resource: resource, Action: "read"},
		"PUT:" + base + "/exposure/:capabilityID":    {Resource: resource, Action: "write"},
		"GET:" + base + "/quotas/:capabilityID":      {Resource: resource, Action: "read"},
		"POST:" + base + "/quotas/:capabilityID":     {Resource: resource, Action: "write"},
		"GET:" + base + "/lifecycle/template":        {Resource: resource, Action: "read"},
		"GET:" + base + "/lifecycle":                 {Resource: resource, Action: "read"},
		"POST:" + base + "/lifecycle":                {Resource: resource, Action: "write"},
		"POST:" + base + "/lifecycle/:planID/status": {Resource: resource, Action: "manage"},
	}
}
