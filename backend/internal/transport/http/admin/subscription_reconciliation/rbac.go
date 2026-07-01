package subscription_reconciliation

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level permissions for subscription reconciliation admin APIs.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/subscription-reconciliation"
	res := func(suffix string) string {
		return "com.powerx.plugins.ecommerce:" + suffix
	}
	return map[string]authx.Permission{
		"GET:" + base + "/batches":            {Resource: res("subscription.reconciliation.batch"), Action: "read"},
		"POST:" + base + "/batches":           {Resource: res("subscription.reconciliation.batch"), Action: "manage"},
		"GET:" + base + "/batches/:id/deltas": {Resource: res("subscription.reconciliation.delta"), Action: "read"},
		"POST:" + base + "/deltas/:id/tasks":  {Resource: res("subscription.reconciliation.task"), Action: "manage"},
		"POST:" + base + "/deltas/:id/adjust": {Resource: res("subscription.reconciliation.delta"), Action: "manage"},
		"POST:" + base + "/tasks/:id/close":   {Resource: res("subscription.reconciliation.task"), Action: "manage"},
		"POST:" + base + "/governance/run":    {Resource: res("subscription.reconciliation.governance"), Action: "manage"},
		"GET:" + base + "/dashboard":          {Resource: res("subscription.reconciliation.dashboard"), Action: "read"},
		"GET:" + base + "/dashboard/export":   {Resource: res("subscription.reconciliation.dashboard"), Action: "export"},
	}
}
