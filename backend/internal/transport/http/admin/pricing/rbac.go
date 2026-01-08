package pricing

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route level permissions for pricing/pricebook API.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/pricing"
	res := func(suffix string) string {
		return "com.powerx.plugin.ecommerce:" + suffix
	}
	return map[string]authx.Permission{
		"GET:" + base + "/pricebooks":                                           {Resource: res("pricing.pricebook"), Action: "read"},
		"POST:" + base + "/pricebooks":                                          {Resource: res("pricing.pricebook"), Action: "manage"},
		"PATCH:" + base + "/pricebooks/:pricebookId":                            {Resource: res("pricing.pricebook"), Action: "manage"},
		"POST:" + base + "/pricebooks/:pricebookId/versions":                    {Resource: res("pricing.pricebook"), Action: "manage"},
		"PUT:" + base + "/pricebooks/:pricebookId/versions/:versionId/items":    {Resource: res("pricing.pricebook"), Action: "manage"},
		"POST:" + base + "/pricebooks/:pricebookId/versions/:versionId/publish": {Resource: res("pricing.pricebook"), Action: "publish"},
		"POST:" + base + "/pricebooks/:pricebookId/versions/:versionId/archive": {Resource: res("pricing.pricebook"), Action: "publish"},
	}
}
