package fulfillment

import (
	"context"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

func RequireTenantUUID(ctx context.Context) (string, error) {
	return AuthX.RequireTenantUUID(ctx)
}
