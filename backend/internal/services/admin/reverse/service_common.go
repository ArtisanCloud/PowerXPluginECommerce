package reverse

import (
	"context"
	"encoding/json"
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

func jsonBytes(value any, fallback []byte) ([]byte, error) {
	if value == nil {
		return fallback, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fallback, err
	}
	if len(data) == 0 {
		return fallback, nil
	}
	return data, nil
}
