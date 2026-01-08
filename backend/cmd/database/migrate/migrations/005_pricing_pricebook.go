package migrations

import pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"

// PricingPricebookTables enumerates pricing/pricebook tables for automated migrations.
var PricingPricebookTables = []interface{}{
	&pricingModel.Pricebook{},
	&pricingModel.PricebookVersion{},
	&pricingModel.PricebookScope{},
	&pricingModel.PricebookItem{},
	&pricingModel.PricebookAuditLog{},
}
