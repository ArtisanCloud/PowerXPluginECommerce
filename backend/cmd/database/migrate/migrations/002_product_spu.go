package migrations

import productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"

// ProductSPUTables enumerates 002-product-spu-management models.
var ProductSPUTables = []interface{}{
	&productmodel.SPU{},
	&productmodel.SPUVersion{},
	&productmodel.SPULocale{},
	&productmodel.ChannelVisibility{},
	&productmodel.SubscriptionPlan{},
	&productmodel.SubscriptionPlanBenefit{},
	&productmodel.SPUImportTask{},
	&productmodel.SPUExportTask{},
	&productmodel.SPUApprovalRecord{},
	&productmodel.SPUAuditLog{},
}
