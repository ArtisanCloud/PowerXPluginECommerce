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
		"POST:" + base + "/templates/:id/quote":   {Resource: res("logistics.template"), Action: "read"},
		"GET:" + base + "/routing/rules":          {Resource: res("logistics.routing"), Action: "read"},
		"POST:" + base + "/routing/rules":         {Resource: res("logistics.routing"), Action: "manage"},
		"PATCH:" + base + "/routing/rules/:id":    {Resource: res("logistics.routing"), Action: "manage"},
		"DELETE:" + base + "/routing/rules/:id":   {Resource: res("logistics.routing"), Action: "manage"},
		"POST:" + base + "/routing/preview":       {Resource: res("logistics.routing"), Action: "read"},
		"GET:" + base + "/eta":                    {Resource: res("logistics.eta"), Action: "read"},
		"GET:" + base + "/eta/:waybill_id":        {Resource: res("logistics.eta"), Action: "read"},
		"POST:" + base + "/waybills":              {Resource: res("logistics.waybill"), Action: "manage"},
		"GET:" + base + "/waybills/:id":           {Resource: res("logistics.waybill"), Action: "read"},
		"POST:" + base + "/waybills/:id/track":    {Resource: res("logistics.waybill"), Action: "manage"},
		"POST:" + base + "/waybills/:id/cancel":   {Resource: res("logistics.waybill"), Action: "manage"},
		"PATCH:" + base + "/waybills/:id/cost":    {Resource: res("logistics.billing"), Action: "manage"},
		"GET:" + base + "/billing/summary":        {Resource: res("logistics.billing"), Action: "read"},
		"GET:" + base + "/billing/export":         {Resource: res("logistics.billing"), Action: "export"},
		"GET:" + base + "/billing/cases":          {Resource: res("logistics.billing.case"), Action: "read"},
		"POST:" + base + "/billing/cases":         {Resource: res("logistics.billing.case"), Action: "manage"},
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
		"POST:" + base + "/webhook":             {Resource: res("logistics.webhook"), Action: "manage"},
	}
}
