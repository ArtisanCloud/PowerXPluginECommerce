package migrations

import productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"

// ProductCategoryTables enumerates 004-product-categories models.
var ProductCategoryTables = []interface{}{
	&productcategory.ProductCategory{},
	&productcategory.CategoryLocale{},
	&productcategory.CategoryTemplate{},
	&productcategory.CategoryTemplateField{},
	&productcategory.CategoryTemplateVersion{},
	&productcategory.CategoryMapping{},
	&productcategory.CategoryPermission{},
}
