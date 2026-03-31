package after_sales

import (
	"context"
	"strings"

	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

func withTenantContext(ctx context.Context, tenantUUID string) context.Context {
	if ctx == nil {
		return nil
	}
	if strings.TrimSpace(tenantUUID) == "" {
		return ctx
	}
	return AuthX.ContextWithTenantUUID(ctx, strings.TrimSpace(tenantUUID))
}
