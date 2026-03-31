package after_sales

import (
	"strings"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for after-sales admin APIs.
func RBACEntries(prefix string) map[string]AuthX.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/after-sales"
	res := func(suffix string) string {
		return "com.powerx.plugins.ecommerce:" + suffix
	}
	return map[string]AuthX.Permission{
		"GET:" + base + "/cases":                        {Resource: res("after_sales.case"), Action: "read"},
		"GET:" + base + "/cases/:id":                    {Resource: res("after_sales.case"), Action: "read"},
		"POST:" + base + "/cases/:id/accept":            {Resource: res("after_sales.case"), Action: "manage"},
		"POST:" + base + "/cases/:id/review":            {Resource: res("after_sales.case"), Action: "manage"},
		"POST:" + base + "/cases/:id/approve":           {Resource: res("after_sales.case"), Action: "manage"},
		"POST:" + base + "/cases/:id/reject":            {Resource: res("after_sales.case"), Action: "manage"},
		"POST:" + base + "/cases/:id/complete":          {Resource: res("after_sales.case"), Action: "manage"},
		"POST:" + base + "/cases/:id/close":             {Resource: res("after_sales.case"), Action: "manage"},
		"GET:" + base + "/dashboard":                    {Resource: res("after_sales.dashboard"), Action: "read"},
		"POST:" + base + "/cases/:id/reverse-logistics": {Resource: res("after_sales.reverse_link"), Action: "manage"},
	}
}
