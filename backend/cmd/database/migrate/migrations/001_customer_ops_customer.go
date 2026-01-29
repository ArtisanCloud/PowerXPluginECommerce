package migrations

import customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"

// CustomerOpsCustomerTables enumerates 001-customer-ops-customer models.
var CustomerOpsCustomerTables = []interface{}{
	&customermodel.Customer{},
	&customermodel.CustomerAccount{},
	&customermodel.CustomerIdentity{},
}
