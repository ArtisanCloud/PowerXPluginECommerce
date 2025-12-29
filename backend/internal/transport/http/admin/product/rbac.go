package product

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares product SKU route permissions.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/product"
	skuBase := base + "/skus"
	spuBase := base + "/spus"
	resource := func(suffix string) string {
		return "com.powerx.plugin.ecommerce:" + suffix
	}

	entries := map[string]authx.Permission{
		"GET:" + skuBase:                                   {Resource: resource("product.sku"), Action: "read"},
		"POST:" + skuBase:                                  {Resource: resource("product.sku"), Action: "manage"},
		"PATCH:" + skuBase + "/:id":                        {Resource: resource("product.sku"), Action: "manage"},
		"DELETE:" + skuBase + "/:id":                       {Resource: resource("product.sku"), Action: "manage"},
		"POST:" + skuBase + "/bulk-tasks":                  {Resource: resource("product.sku.bulk"), Action: "manage"},
		"GET:" + skuBase + "/bulk-tasks/:taskId":           {Resource: resource("product.sku.bulk"), Action: "read"},
		"POST:" + skuBase + "/bulk-tasks/:taskId/approval": {Resource: resource("product.sku.bulk"), Action: "manage"},
		"POST:" + skuBase + "/bulk-tasks/:taskId/retry":    {Resource: resource("product.sku.bulk"), Action: "manage"},
		"POST:" + skuBase + "/import":                      {Resource: resource("product.sku.bulk"), Action: "manage"},
		"POST:" + skuBase + "/export":                      {Resource: resource("product.sku.bulk"), Action: "manage"},
		"GET:" + skuBase + "/:id/inventory":                {Resource: resource("product.sku.inventory"), Action: "read"},
		"GET:" + skuBase + "/:id/channels":                 {Resource: resource("product.sku.channel"), Action: "read"},
		"POST:" + skuBase + "/:id/channels":                {Resource: resource("product.sku.channel"), Action: "manage"},
		"POST:" + skuBase + "/:id/channels/publish":        {Resource: resource("product.sku.channel"), Action: "manage"},
		"GET:" + skuBase + "/:id/serials":                  {Resource: resource("product.sku.serial"), Action: "read"},
		"POST:" + skuBase + "/:id/serials":                 {Resource: resource("product.sku.serial"), Action: "manage"},
		"POST:" + skuBase + "/:id/barcodes":                {Resource: resource("product.sku.barcode"), Action: "manage"},
		"POST:" + spuBase + "/:id/skus/generate":           {Resource: resource("product.sku"), Action: "manage"},
	}
	return entries
}
