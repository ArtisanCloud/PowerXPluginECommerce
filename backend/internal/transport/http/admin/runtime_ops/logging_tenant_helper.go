package runtime_ops

import (
	"strings"

	httpmiddleware "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func resolvePolicyTenant(c *gin.Context, requested string) (tenantUUID string, mismatch bool) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		if headerTenant := strings.TrimSpace(c.GetHeader("tenant_uuid")); headerTenant != "" {
			requested = headerTenant
		} else if queryTenant := strings.TrimSpace(c.Query("tenant_uuid")); queryTenant != "" {
			requested = queryTenant
		}
	}

	ctxTenant := ""
	if id, ok := httpmiddleware.TenantUUIDFromContext(c); ok {
		ctxTenant = strings.TrimSpace(id)
	}

	if requested != "" && ctxTenant != "" && !strings.EqualFold(requested, ctxTenant) {
		return ctxTenant, true
	}
	if ctxTenant != "" {
		return ctxTenant, false
	}
	return requested, false
}
