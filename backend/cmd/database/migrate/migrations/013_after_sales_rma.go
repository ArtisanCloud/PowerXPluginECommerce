package migrations

import (
	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
)

// AfterSalesRMATables enumerates 012-after-sales-rma models.
var AfterSalesRMATables = []interface{}{
	&AfterSalesModel.AfterSaleCase{},
	&AfterSalesModel.AfterSaleTimeline{},
	&AfterSalesModel.AfterSaleEvidence{},
	&AfterSalesModel.AfterSaleDecision{},
	&AfterSalesModel.ReturnLogisticsLink{},
}
