package migrations

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	orderModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
)

// OrderTables enumerates first-party order tables for automated migrations.
var OrderTables = []interface{}{
	&orderModel.Order{},
	&orderModel.OrderItem{},
	&orderModel.OrderEvent{},
	&models.OrderBenefitReview{},
}
