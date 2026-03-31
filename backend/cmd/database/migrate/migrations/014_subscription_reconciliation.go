package migrations

import (
	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
)

// SubscriptionReconciliationTables defines 013-subscription-reconciliation tables.
var SubscriptionReconciliationTables = []interface{}{
	&SubscriptionReconciliationModel.ReconciliationBatch{},
	&SubscriptionReconciliationModel.ReconciliationDelta{},
	&SubscriptionReconciliationModel.DeltaTask{},
	&SubscriptionReconciliationModel.RenewalGovernancePolicy{},
	&SubscriptionReconciliationModel.RenewalExecutionLog{},
}
