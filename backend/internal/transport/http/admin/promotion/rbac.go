package promotion

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for promotion admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/promotions"
	res := func(suffix string) string { return "com.powerx.plugins.ecommerce:" + suffix }
	return map[string]authx.Permission{
		"GET:" + base:                     {Resource: res("pricing.promotion"), Action: "read"},
		"POST:" + base:                    {Resource: res("pricing.promotion"), Action: "manage"},
		"PATCH:" + base + "/:id":          {Resource: res("pricing.promotion"), Action: "manage"},
		"POST:" + base + "/:id/activate":  {Resource: res("pricing.promotion"), Action: "publish"},
		"POST:" + base + "/:id/pause":     {Resource: res("pricing.promotion"), Action: "publish"},
		"POST:" + base + "/:id/clone":     {Resource: res("pricing.promotion"), Action: "manage"},
		"GET:" + base + "/:id/audit-logs": {Resource: res("pricing.promotion.audit"), Action: "read"},
	}
}
