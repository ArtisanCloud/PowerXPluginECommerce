package payments

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for payment admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/payments"
	res := func() string { return "com.powerx.plugin.ecommerce:payments" }
	return map[string]authx.Permission{
		"GET:" + base + "/providers":        {Resource: res(), Action: "read"},
		"POST:" + base + "/providers":       {Resource: res(), Action: "manage"},
		"PATCH:" + base + "/providers/:id":  {Resource: res(), Action: "manage"},
		"POST:" + base + "/providers/:id/test": {Resource: res(), Action: "manage"},
		"GET:" + base + "/transactions":     {Resource: res(), Action: "read"},
		"GET:" + base + "/transactions/:id": {Resource: res(), Action: "read"},
		"POST:" + base + "/transactions/:id/refund": {Resource: res(), Action: "refund"},
		"GET:" + base + "/risk-events":       {Resource: res(), Action: "risk"},
		"GET:" + base + "/split-rules":       {Resource: res(), Action: "read"},
		"GET:" + base + "/split-results":     {Resource: res(), Action: "read"},
		"GET:" + base + "/reconciliations":   {Resource: res(), Action: "read"},
		"POST:" + base + "/reconciliations":  {Resource: res(), Action: "manage"},
		"GET:" + base + "/reconciliations/:id/items": {Resource: res(), Action: "read"},
		"POST:" + base + "/reconciliations/:id/items/:itemId/resolve": {Resource: res(), Action: "manage"},
		"GET:" + base + "/manual-payments":      {Resource: res(), Action: "read"},
		"GET:" + base + "/manual-payments/logs": {Resource: res(), Action: "read"},
		"POST:" + base + "/manual-payments":     {Resource: res(), Action: "manage"},
		"POST:" + base + "/manual-payments/:id/approve": {Resource: res(), Action: "manage"},
		"POST:" + base + "/manual-payments/:id/reject":  {Resource: res(), Action: "manage"},
	}
}
