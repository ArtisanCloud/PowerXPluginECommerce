package logistics

import (
	"strings"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for logistics admin APIs.
func RBACEntries(prefix string) map[string]AuthX.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/logistics"
	res := func(suffix string) string {
		return "com.powerx.plugins.ecommerce:" + suffix
	}
	return map[string]AuthX.Permission{
		"GET:" + base + "/carriers":               {Resource: res("logistics.carrier"), Action: "read"},
		"POST:" + base + "/carriers":              {Resource: res("logistics.carrier"), Action: "manage"},
		"PATCH:" + base + "/carriers/:id":         {Resource: res("logistics.carrier"), Action: "manage"},
		"POST:" + base + "/carriers/:id/test":     {Resource: res("logistics.carrier"), Action: "manage"},
		"GET:" + base + "/templates":              {Resource: res("logistics.template"), Action: "read"},
		"POST:" + base + "/templates":             {Resource: res("logistics.template"), Action: "manage"},
		"PATCH:" + base + "/templates/:id":        {Resource: res("logistics.template"), Action: "manage"},
		"POST:" + base + "/templates/:id/publish": {Resource: res("logistics.template"), Action: "manage"},
		"POST:" + base + "/waybills":              {Resource: res("logistics.waybill"), Action: "manage"},
		"GET:" + base + "/waybills/:id":           {Resource: res("logistics.waybill"), Action: "read"},
		"POST:" + base + "/waybills/:id/track":    {Resource: res("logistics.waybill"), Action: "manage"},
		"POST:" + base + "/waybills/:id/cancel":   {Resource: res("logistics.waybill"), Action: "manage"},
		"POST:" + base + "/webhook":               {Resource: res("logistics.webhook"), Action: "manage"},
	}
}
