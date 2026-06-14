package migrations

import promotionModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"

// PromotionTables enumerates promotion tables for automated migrations.
var PromotionTables = []interface{}{
	&promotionModel.Campaign{},
	&promotionModel.AuditLog{},
	&promotionModel.OrderSnapshot{},
}
