package migrations

import cartModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/cart"

// CartTables enumerates cart tables for automated migrations.
var CartTables = []interface{}{
	&cartModel.Cart{},
}
