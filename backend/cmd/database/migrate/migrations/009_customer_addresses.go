package migrations

import customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"

// CustomerAddressTables enumerates address book tables for automated migrations.
var CustomerAddressTables = []interface{}{
	&customermodel.CustomerAddress{},
}
