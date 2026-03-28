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
		"GET:" + base + "/waves":                {Resource: res("fulfillment.wave"), Action: "read"},
		"POST:" + base + "/waves":               {Resource: res("fulfillment.wave"), Action: "manage"},
		"GET:" + base + "/waves/:id":            {Resource: res("fulfillment.wave"), Action: "read"},
		"PATCH:" + base + "/waves/:id/advance":  {Resource: res("fulfillment.wave"), Action: "manage"},
		"PATCH:" + base + "/waves/:id/tasks/:task_id/reassign": {
			Resource: res("fulfillment.wave"), Action: "manage",
		},
		"GET:" + base + "/wave-strategies":          {Resource: res("fulfillment.wave.strategy"), Action: "read"},
		"POST:" + base + "/wave-strategies":         {Resource: res("fulfillment.wave.strategy"), Action: "manage"},
		"POST:" + base + "/wave-strategies/preview": {Resource: res("fulfillment.wave.strategy"), Action: "read"},
		"GET:" + base + "/exceptions":               {Resource: res("fulfillment.exception"), Action: "read"},
		"POST:" + base + "/exceptions":              {Resource: res("fulfillment.exception"), Action: "manage"},
		"GET:" + base + "/warehouse/outbounds":      {Resource: res("fulfillment.warehouse"), Action: "read"},
		"POST:" + base + "/warehouse/outbounds":     {Resource: res("fulfillment.warehouse"), Action: "manage"},
		"POST:" + base + "/warehouse/outbounds/:id/execute": {
			Resource: res("fulfillment.warehouse"), Action: "manage",
		},
		"POST:" + base + "/warehouse/outbounds/:id/rollback": {
			Resource: res("fulfillment.warehouse"), Action: "manage",
		},
	}
}
