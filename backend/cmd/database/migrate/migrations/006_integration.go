package migrations

import integrationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/integration"

// IntegrationTables enumerates integration tables required by runtime services (eg idempotency).
var IntegrationTables = []interface{}{
	&integrationModel.IdempotencyRecord{},
}

