package migrations

import models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"

var PaymentTables = []interface{}{
	&models.PaymentProvider{},
	&models.PaymentTransaction{},
	&models.PaymentRefund{},
	&models.PaymentRiskEvent{},
	&models.PaymentReconciliation{},
	&models.PaymentReconciliationItem{},
	&models.PaymentSplitRule{},
	&models.PaymentSplitResult{},
	&models.PaymentManualReview{},
	&models.PaymentManualReviewLog{},
}
