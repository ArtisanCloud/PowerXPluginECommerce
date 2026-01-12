package migrations

import (
	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
)

// ProductSkuTables enumerates all SKU-related tables for automated migrations.
var ProductSkuTables = []interface{}{
	&productspecmodel.ProductSpecGroup{},
	&productspecmodel.ProductSpecOption{},
	&productskumodel.ProductSKU{},
	&productskumodel.ProductSKUAttribute{},
	&productskumodel.ProductSKUChannel{},
	&productskumodel.ProductSKUInventory{},
	&productskumodel.ProductSKUMedia{},
	&productskumodel.ProductSKUBulkTask{},
	&productskumodel.ProductSKUBulkTaskItem{},
	&productskumodel.ProductSKUSerialRecord{},
	&productskumodel.ProductSKUAuditLog{},
}
