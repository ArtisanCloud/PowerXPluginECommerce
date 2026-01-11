package product

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route-level RBAC mappings for product admin APIs (SKU/SPU/categories/templates).
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/product"
	skuBase := base + "/skus"
	spuBase := base + "/spus"
	categoryBase := base + "/categories"
	templateBase := base + "/category-templates"
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
		"POST:" + skuBase + "/:id/inventory/adjust":        {Resource: resource("product.sku.inventory"), Action: "manage"},
		"GET:" + skuBase + "/:id/channels":                 {Resource: resource("product.sku.channel"), Action: "read"},
		"POST:" + skuBase + "/:id/channels":                {Resource: resource("product.sku.channel"), Action: "manage"},
		"POST:" + skuBase + "/:id/channels/publish":        {Resource: resource("product.sku.channel"), Action: "manage"},
		"GET:" + skuBase + "/:id/serials":                  {Resource: resource("product.sku.serial"), Action: "read"},
		"POST:" + skuBase + "/:id/serials":                 {Resource: resource("product.sku.serial"), Action: "manage"},
		"POST:" + skuBase + "/:id/barcodes":                {Resource: resource("product.sku.barcode"), Action: "manage"},
		"POST:" + spuBase + "/:id/skus/generate":           {Resource: resource("product.sku"), Action: "manage"},

		// Category CRUD.
		"GET:" + categoryBase:                   {Resource: resource("product.category"), Action: "read"},
		"GET:" + categoryBase + "/tree":         {Resource: resource("product.category"), Action: "read"},
		"POST:" + categoryBase:                  {Resource: resource("product.category"), Action: "manage"},
		"PATCH:" + categoryBase + "/:id":        {Resource: resource("product.category"), Action: "manage"},
		"DELETE:" + categoryBase + "/:id":       {Resource: resource("product.category"), Action: "manage"},
		"POST:" + categoryBase + "/:id/move":    {Resource: resource("product.category"), Action: "manage"},
		"PATCH:" + categoryBase + "/:id/status": {Resource: resource("product.category"), Action: "manage"},
		"GET:" + categoryBase + "/:id/audit":    {Resource: resource("product.category"), Action: "read"},

		// Category mappings.
		"GET:" + categoryBase + "/:id/mappings":  {Resource: resource("product.category.mapping"), Action: "read"},
		"POST:" + categoryBase + "/:id/mappings": {Resource: resource("product.category.mapping"), Action: "manage"},

		// Import/export.
		"POST:" + categoryBase + "/import": {Resource: resource("product.category.import"), Action: "manage"},
		"POST:" + categoryBase + "/export": {Resource: resource("product.category.import"), Action: "read"},

		// Category templates.
		"GET:" + templateBase:                    {Resource: resource("product.category.template"), Action: "read"},
		"POST:" + templateBase:                   {Resource: resource("product.category.template"), Action: "manage"},
		"PATCH:" + templateBase + "/:id":         {Resource: resource("product.category.template"), Action: "manage"},
		"POST:" + templateBase + "/:id/publish":  {Resource: resource("product.category.template"), Action: "manage"},
		"POST:" + templateBase + "/:id/rollback": {Resource: resource("product.category.template"), Action: "manage"},
	}
	return entries
}
