package coupon

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for coupon admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/coupons"
	res := func(suffix string) string { return "com.powerx.plugins.ecommerce:" + suffix }
	return map[string]authx.Permission{
		"GET:" + base + "/templates":       {Resource: res("pricing.coupon.template"), Action: "read"},
		"POST:" + base + "/templates":      {Resource: res("pricing.coupon.template"), Action: "manage"},
		"PATCH:" + base + "/templates/:id": {Resource: res("pricing.coupon.template"), Action: "manage"},
		"POST:" + base + "/issues":         {Resource: res("pricing.coupon.issue"), Action: "manage"},
		"GET:" + base + "/assets":          {Resource: res("pricing.coupon.asset"), Action: "read"},
		"GET:" + base + "/usage-logs":      {Resource: res("pricing.coupon.usage"), Action: "read"},
	}
}
