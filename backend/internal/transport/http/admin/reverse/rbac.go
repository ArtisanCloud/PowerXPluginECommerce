package reverse

import (
	"strings"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for reverse logistics admin APIs.
func RBACEntries(prefix string) map[string]AuthX.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/reverse"
	res := func(suffix string) string {
		return "com.powerx.plugins.ecommerce:" + suffix
	}
	return map[string]AuthX.Permission{
		"GET:" + base + "/waybills":                       {Resource: res("reverse.waybill"), Action: "read"},
		"POST:" + base + "/waybills":                      {Resource: res("reverse.waybill"), Action: "manage"},
		"GET:" + base + "/waybills/:id":                   {Resource: res("reverse.waybill"), Action: "read"},
		"POST:" + base + "/waybills/:id/track":            {Resource: res("reverse.waybill"), Action: "manage"},
		"POST:" + base + "/waybills/:id/warehouse-result": {Resource: res("reverse.waybill"), Action: "manage"},
		"GET:" + base + "/inspection/rules":               {Resource: res("reverse.inspection"), Action: "read"},
		"POST:" + base + "/inspection/rules":              {Resource: res("reverse.inspection"), Action: "manage"},
		"POST:" + base + "/waybills/:id/inspection":       {Resource: res("reverse.inspection"), Action: "manage"},
	}
}
