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
		"GET:" + base + "/carriers":                    {Resource: res("logistics.carrier"), Action: "read"},
		"POST:" + base + "/carriers":                   {Resource: res("logistics.carrier"), Action: "manage"},
		"PATCH:" + base + "/carriers/:id":              {Resource: res("logistics.carrier"), Action: "manage"},
		"POST:" + base + "/carriers/:id/test":          {Resource: res("logistics.carrier"), Action: "manage"},
		"GET:" + base + "/carriers/profiles":           {Resource: res("logistics.carrier_profile"), Action: "read"},
		"POST:" + base + "/carriers/profiles/evaluate": {Resource: res("logistics.carrier_profile"), Action: "manage"},
		"POST:" + base + "/carriers/profiles/:id/confirm-rating": {
			Resource: res("logistics.carrier_profile"), Action: "manage",
		},
		"POST:" + base + "/carriers/profiles/:id/retire": {
			Resource: res("logistics.carrier_profile"), Action: "manage",
		},
		"POST:" + base + "/carriers/profiles/:id/restore": {
			Resource: res("logistics.carrier_profile"), Action: "manage",
		},
		"GET:" + base + "/templates":                     {Resource: res("logistics.template"), Action: "read"},
		"POST:" + base + "/templates":                    {Resource: res("logistics.template"), Action: "manage"},
		"PATCH:" + base + "/templates/:id":               {Resource: res("logistics.template"), Action: "manage"},
		"POST:" + base + "/templates/:id/publish":        {Resource: res("logistics.template"), Action: "manage"},
		"POST:" + base + "/templates/:id/quote":          {Resource: res("logistics.template"), Action: "read"},
		"GET:" + base + "/routing/rules":                 {Resource: res("logistics.routing"), Action: "read"},
		"POST:" + base + "/routing/rules":                {Resource: res("logistics.routing"), Action: "manage"},
		"PATCH:" + base + "/routing/rules/:id":           {Resource: res("logistics.routing"), Action: "manage"},
		"DELETE:" + base + "/routing/rules/:id":          {Resource: res("logistics.routing"), Action: "manage"},
		"POST:" + base + "/routing/preview":              {Resource: res("logistics.routing"), Action: "read"},
		"GET:" + base + "/redelivery/tasks":              {Resource: res("logistics.redelivery"), Action: "read"},
		"POST:" + base + "/redelivery/tasks/initiate":    {Resource: res("logistics.redelivery"), Action: "manage"},
		"POST:" + base + "/redelivery/tasks/:id/address": {Resource: res("logistics.redelivery"), Action: "manage"},
		"POST:" + base + "/redelivery/tasks/:id/redispatch": {
			Resource: res("logistics.redelivery"), Action: "manage",
		},
		"POST:" + base + "/redelivery/tasks/:id/close": {Resource: res("logistics.redelivery"), Action: "manage"},
		"GET:" + base + "/risk/rules":                  {Resource: res("logistics.risk"), Action: "read"},
		"POST:" + base + "/risk/rules":                 {Resource: res("logistics.risk"), Action: "manage"},
		"PATCH:" + base + "/risk/rules/:id":            {Resource: res("logistics.risk"), Action: "manage"},
		"GET:" + base + "/risk/blacklist":              {Resource: res("logistics.risk"), Action: "read"},
		"POST:" + base + "/risk/blacklist":             {Resource: res("logistics.risk"), Action: "manage"},
		"PATCH:" + base + "/risk/blacklist/:id":        {Resource: res("logistics.risk"), Action: "manage"},
		"GET:" + base + "/risk/hits":                   {Resource: res("logistics.risk"), Action: "read"},
		"POST:" + base + "/risk/evaluate":              {Resource: res("logistics.risk"), Action: "manage"},
		"POST:" + base + "/risk/hits/:id/release":      {Resource: res("logistics.risk"), Action: "manage"},
		"GET:" + base + "/eta":                         {Resource: res("logistics.eta"), Action: "read"},
		"GET:" + base + "/eta/:waybill_id":             {Resource: res("logistics.eta"), Action: "read"},
		"POST:" + base + "/waybills":                   {Resource: res("logistics.waybill"), Action: "manage"},
		"GET:" + base + "/waybills/:id":                {Resource: res("logistics.waybill"), Action: "read"},
		"POST:" + base + "/waybills/:id/track":         {Resource: res("logistics.waybill"), Action: "manage"},
		"POST:" + base + "/waybills/:id/sync-track":    {Resource: res("logistics.waybill"), Action: "manage"},
		"POST:" + base + "/waybills/:id/cancel":        {Resource: res("logistics.waybill"), Action: "manage"},
		"PATCH:" + base + "/waybills/:id/cost":         {Resource: res("logistics.billing"), Action: "manage"},
		"GET:" + base + "/billing/summary":             {Resource: res("logistics.billing"), Action: "read"},
		"GET:" + base + "/billing/export":              {Resource: res("logistics.billing"), Action: "export"},
		"GET:" + base + "/billing/cases":               {Resource: res("logistics.billing.case"), Action: "read"},
		"POST:" + base + "/billing/cases":              {Resource: res("logistics.billing.case"), Action: "manage"},
		"PATCH:" + base + "/billing/cases/:id/transition": {
			Resource: res("logistics.billing.case"), Action: "manage",
		},
		"GET:" + base + "/notifications/templates":  {Resource: res("logistics.notification"), Action: "read"},
		"POST:" + base + "/notifications/templates": {Resource: res("logistics.notification"), Action: "manage"},
		"GET:" + base + "/notifications/records":    {Resource: res("logistics.notification"), Action: "read"},
		"POST:" + base + "/notifications/send":      {Resource: res("logistics.notification"), Action: "manage"},
		"POST:" + base + "/notifications/records/:id/retry": {
			Resource: res("logistics.notification"), Action: "manage",
		},
		"GET:" + base + "/sla/dashboard":        {Resource: res("logistics.sla"), Action: "read"},
		"GET:" + base + "/labels/prints":        {Resource: res("logistics.label_print"), Action: "read"},
		"POST:" + base + "/labels/prints":       {Resource: res("logistics.label_print"), Action: "manage"},
		"POST:" + base + "/labels/prints/retry": {Resource: res("logistics.label_print"), Action: "manage"},
		"GET:" + base + "/tracking-sync/jobs":   {Resource: res("logistics.tracking_sync"), Action: "read"},
		"POST:" + base + "/tracking-sync/jobs":  {Resource: res("logistics.tracking_sync"), Action: "manage"},
		"POST:" + base + "/tracking-sync/jobs/:id/cancel": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"POST:" + base + "/tracking-sync/jobs/:id/retry": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"GET:" + base + "/tracking-sync/schedules": {Resource: res("logistics.tracking_sync"), Action: "read"},
		"POST:" + base + "/tracking-sync/schedules": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"PATCH:" + base + "/tracking-sync/schedules/:id": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"POST:" + base + "/tracking-sync/schedules/:id/toggle": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"POST:" + base + "/tracking-sync/schedules/:id/trigger": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"POST:" + base + "/tracking-sync/schedules/run-due": {
			Resource: res("logistics.tracking_sync"), Action: "manage",
		},
		"GET:" + base + "/gateway/health": {Resource: res("logistics.gateway"), Action: "read"},
		"GET:" + base + "/gateway/costs":  {Resource: res("logistics.gateway"), Action: "read"},
		"GET:" + base + "/gateway/cost-alerts": {
			Resource: res("logistics.gateway"), Action: "read",
		},
		"GET:" + base + "/gateway/failures": {
			Resource: res("logistics.gateway"), Action: "read",
		},
		"POST:" + base + "/gateway/failures/ingest": {
			Resource: res("logistics.gateway"), Action: "manage",
		},
		"POST:" + base + "/gateway/failures/:id/compensate": {
			Resource: res("logistics.gateway"), Action: "manage",
		},
		"GET:" + base + "/exceptions/orchestration/rules": {Resource: res("logistics.orchestration"), Action: "read"},
		"POST:" + base + "/exceptions/orchestration/rules": {
			Resource: res("logistics.orchestration"), Action: "manage",
		},
		"PATCH:" + base + "/exceptions/orchestration/rules/:id": {
			Resource: res("logistics.orchestration"), Action: "manage",
		},
		"GET:" + base + "/exceptions/orchestration/runs": {Resource: res("logistics.orchestration"), Action: "read"},
		"POST:" + base + "/exceptions/orchestration/execute": {
			Resource: res("logistics.orchestration"), Action: "manage",
		},
		"POST:" + base + "/address-validation/check": {
			Resource: res("logistics.address_validation"), Action: "manage",
		},
		"GET:" + base + "/address-validation/records": {
			Resource: res("logistics.address_validation"), Action: "read",
		},
		"GET:" + base + "/routing/optimizer/strategy": {
			Resource: res("logistics.routing_optimizer"), Action: "read",
		},
		"POST:" + base + "/routing/optimizer/strategy": {
			Resource: res("logistics.routing_optimizer"), Action: "manage",
		},
		"POST:" + base + "/routing/optimizer/simulate": {
			Resource: res("logistics.routing_optimizer"), Action: "read",
		},
		"GET:" + base + "/settlement/batches": {
			Resource: res("logistics.settlement"), Action: "read",
		},
		"POST:" + base + "/settlement/batches": {
			Resource: res("logistics.settlement"), Action: "manage",
		},
		"GET:" + base + "/settlement/diffs": {
			Resource: res("logistics.settlement"), Action: "read",
		},
		"POST:" + base + "/settlement/diffs/:id/handle": {
			Resource: res("logistics.settlement"), Action: "manage",
		},
		"POST:" + base + "/settlement/batches/:id/confirm": {
			Resource: res("logistics.settlement"), Action: "manage",
		},
		"GET:" + base + "/reconciliation/batches": {
			Resource: res("logistics.reconciliation"), Action: "read",
		},
		"POST:" + base + "/reconciliation/batches": {
			Resource: res("logistics.reconciliation"), Action: "manage",
		},
		"GET:" + base + "/reconciliation/records": {
			Resource: res("logistics.reconciliation"), Action: "read",
		},
		"GET:" + base + "/reconciliation/cases": {
			Resource: res("logistics.reconciliation"), Action: "read",
		},
		"POST:" + base + "/reconciliation/cases/:id/handle": {
			Resource: res("logistics.reconciliation"), Action: "manage",
		},
		"GET:" + base + "/control-tower/overview": {
			Resource: res("logistics.control_tower"), Action: "read",
		},
		"GET:" + base + "/control-tower/drilldown": {
			Resource: res("logistics.control_tower"), Action: "read",
		},
		"GET:" + base + "/control-tower/subscriptions": {
			Resource: res("logistics.control_tower"), Action: "read",
		},
		"POST:" + base + "/control-tower/subscriptions": {
			Resource: res("logistics.control_tower"), Action: "manage",
		},
		"PATCH:" + base + "/control-tower/subscriptions/:id": {
			Resource: res("logistics.control_tower"), Action: "manage",
		},
		"GET:" + base + "/allocation/plans": {
			Resource: res("logistics.allocation"), Action: "read",
		},
		"POST:" + base + "/allocation/plans": {
			Resource: res("logistics.allocation"), Action: "manage",
		},
		"PATCH:" + base + "/allocation/plans/:id": {
			Resource: res("logistics.allocation"), Action: "manage",
		},
		"POST:" + base + "/allocation/allocate": {
			Resource: res("logistics.allocation"), Action: "manage",
		},
		"POST:" + base + "/allocation/override": {
			Resource: res("logistics.allocation"), Action: "manage",
		},
		"GET:" + base + "/allocation/forecasts": {
			Resource: res("logistics.capacity_forecast"), Action: "read",
		},
		"POST:" + base + "/allocation/forecasts/generate": {
			Resource: res("logistics.capacity_forecast"), Action: "manage",
		},
		"POST:" + base + "/allocation/forecasts/:id/apply": {
			Resource: res("logistics.capacity_forecast"), Action: "manage",
		},
		"GET:" + base + "/sandbox/scenarios": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "read",
		},
		"POST:" + base + "/sandbox/scenarios": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "manage",
		},
		"PATCH:" + base + "/sandbox/scenarios/:id": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "manage",
		},
		"GET:" + base + "/sandbox/runs": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "read",
		},
		"POST:" + base + "/sandbox/run": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "manage",
		},
		"POST:" + base + "/sandbox/compare": {
			Resource: res("logistics.fulfillment_sandbox"), Action: "read",
		},
		"GET:" + base + "/policy-orchestration/flows": {
			Resource: res("logistics.policy_orchestration"), Action: "read",
		},
		"POST:" + base + "/policy-orchestration/flows": {
			Resource: res("logistics.policy_orchestration"), Action: "manage",
		},
		"PATCH:" + base + "/policy-orchestration/flows/:id": {
			Resource: res("logistics.policy_orchestration"), Action: "manage",
		},
		"GET:" + base + "/policy-orchestration/flows/:id/versions": {
			Resource: res("logistics.policy_orchestration"), Action: "read",
		},
		"POST:" + base + "/policy-orchestration/conflicts/preview": {
			Resource: res("logistics.policy_orchestration"), Action: "read",
		},
		"POST:" + base + "/policy-orchestration/flows/:id/publish": {
			Resource: res("logistics.policy_orchestration"), Action: "manage",
		},
		"POST:" + base + "/policy-orchestration/flows/:id/rollback": {
			Resource: res("logistics.policy_orchestration"), Action: "manage",
		},
		"GET:" + base + "/allocation/interwarehouse/candidates": {
			Resource: res("logistics.interwarehouse"), Action: "read",
		},
		"POST:" + base + "/allocation/interwarehouse/suggest": {
			Resource: res("logistics.interwarehouse"), Action: "manage",
		},
		"POST:" + base + "/allocation/interwarehouse/confirm": {
			Resource: res("logistics.interwarehouse"), Action: "manage",
		},
		"GET:" + base + "/tracking-root-causes/summary": {
			Resource: res("logistics.tracking_root_cause"), Action: "read",
		},
		"GET:" + base + "/tracking-root-causes/items": {
			Resource: res("logistics.tracking_root_cause"), Action: "read",
		},
		"POST:" + base + "/tracking-root-causes/analyze": {
			Resource: res("logistics.tracking_root_cause"), Action: "manage",
		},
		"POST:" + base + "/tracking-root-causes/:id/handle": {
			Resource: res("logistics.tracking_root_cause"), Action: "manage",
		},
		"GET:" + base + "/quality-reports": {
			Resource: res("logistics.quality_audit"), Action: "read",
		},
		"GET:" + base + "/quality-reports/:id": {
			Resource: res("logistics.quality_audit"), Action: "read",
		},
		"POST:" + base + "/quality-reports/generate": {
			Resource: res("logistics.quality_audit"), Action: "manage",
		},
		"GET:" + base + "/quality-reports/export": {
			Resource: res("logistics.quality_audit"), Action: "export",
		},
		"GET:" + base + "/lastmile-recovery/rules": {
			Resource: res("logistics.lastmile_recovery"), Action: "read",
		},
		"POST:" + base + "/lastmile-recovery/rules": {
			Resource: res("logistics.lastmile_recovery"), Action: "manage",
		},
		"PATCH:" + base + "/lastmile-recovery/rules/:id": {
			Resource: res("logistics.lastmile_recovery"), Action: "manage",
		},
		"POST:" + base + "/lastmile-recovery/execute": {
			Resource: res("logistics.lastmile_recovery"), Action: "manage",
		},
		"GET:" + base + "/lastmile-recovery/runs": {
			Resource: res("logistics.lastmile_recovery"), Action: "read",
		},
		"POST:" + base + "/lastmile-recovery/runs/:id/takeover": {
			Resource: res("logistics.lastmile_recovery"), Action: "manage",
		},
		"GET:" + base + "/crossborder/documents": {
			Resource: res("logistics.crossborder"), Action: "read",
		},
		"POST:" + base + "/crossborder/documents": {
			Resource: res("logistics.crossborder"), Action: "manage",
		},
		"PATCH:" + base + "/crossborder/documents/:id": {
			Resource: res("logistics.crossborder"), Action: "manage",
		},
		"POST:" + base + "/crossborder/tax-quote": {
			Resource: res("logistics.crossborder"), Action: "read",
		},
		"GET:" + base + "/crossborder/tracking-maps": {
			Resource: res("logistics.crossborder"), Action: "read",
		},
		"POST:" + base + "/crossborder/tracking-maps": {
			Resource: res("logistics.crossborder"), Action: "manage",
		},
		"PATCH:" + base + "/crossborder/tracking-maps/:id": {
			Resource: res("logistics.crossborder"), Action: "manage",
		},
		"POST:" + base + "/crossborder/tracking-maps/normalize": {
			Resource: res("logistics.crossborder"), Action: "read",
		},
		"GET:" + base + "/customs/rule-packs": {
			Resource: res("logistics.customs_rule"), Action: "read",
		},
		"POST:" + base + "/customs/rule-packs": {
			Resource: res("logistics.customs_rule"), Action: "manage",
		},
		"PATCH:" + base + "/customs/rule-packs/:id": {
			Resource: res("logistics.customs_rule"), Action: "manage",
		},
		"GET:" + base + "/customs/rule-packs/:id/versions": {
			Resource: res("logistics.customs_rule"), Action: "read",
		},
		"POST:" + base + "/customs/rule-packs/:id/versions": {
			Resource: res("logistics.customs_rule"), Action: "manage",
		},
		"POST:" + base + "/customs/precheck": {
			Resource: res("logistics.customs_rule"), Action: "read",
		},
		"GET:" + base + "/customs/compliance-kb/versions": {
			Resource: res("logistics.compliance_kb"), Action: "read",
		},
		"POST:" + base + "/customs/compliance-kb/sync": {
			Resource: res("logistics.compliance_kb"), Action: "manage",
		},
		"POST:" + base + "/customs/compliance-kb/diff": {
			Resource: res("logistics.compliance_kb"), Action: "read",
		},
		"POST:" + base + "/customs/compliance-kb/:id/publish": {
			Resource: res("logistics.compliance_kb"), Action: "manage",
		},
		"GET:" + base + "/kpi-dashboard/overview": {
			Resource: res("logistics.kpi_dashboard"), Action: "read",
		},
		"GET:" + base + "/kpi-dashboard/trends": {
			Resource: res("logistics.kpi_dashboard"), Action: "read",
		},
		"GET:" + base + "/kpi-dashboard/drilldown": {
			Resource: res("logistics.kpi_dashboard"), Action: "read",
		},
		"GET:" + base + "/kpi-dashboard/export": {
			Resource: res("logistics.kpi_dashboard"), Action: "export",
		},
		"GET:" + base + "/slo-guard/policies": {
			Resource: res("logistics.slo_guard"), Action: "read",
		},
		"POST:" + base + "/slo-guard/policies": {
			Resource: res("logistics.slo_guard"), Action: "manage",
		},
		"PATCH:" + base + "/slo-guard/policies/:id": {
			Resource: res("logistics.slo_guard"), Action: "manage",
		},
		"GET:" + base + "/slo-guard/status": {
			Resource: res("logistics.slo_guard"), Action: "read",
		},
		"POST:" + base + "/slo-guard/evaluate": {
			Resource: res("logistics.slo_guard"), Action: "manage",
		},
		"POST:" + base + "/slo-guard/policies/:id/release": {
			Resource: res("logistics.slo_guard"), Action: "manage",
		},
		"POST:" + base + "/webhook": {Resource: res("logistics.webhook"), Action: "manage"},
	}
}
