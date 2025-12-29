package middleware

import (
	"net/http"
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/gin-gonic/gin"
)

// CustomerAuthenticate 校验 mini-app 客户凭证。
func CustomerAuthenticate(authenticator customerauth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticator == nil {
			// 未配置鉴权器时直接放行（用于开发模式）
			c.Next()
			return
		}
		token := extractCustomerToken(c)
		ctx, err := authenticator.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "customer unauthorized"})
			return
		}
		if tenantUUID, ok := TenantUUIDFromContext(c); ok && tenantUUID != "" {
			if !strings.EqualFold(tenantUUID, ctx.TenantUUID) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "tenant mismatch"})
				return
			}
		}
		authx.SetCustomerContext(c, authx.CustomerContext{
			TenantUUID: ctx.TenantUUID,
			CustomerID: ctx.CustomerID,
			Roles:      ctx.Roles,
		})
		c.Next()
	}
}

func extractCustomerToken(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}
	if token := strings.TrimSpace(c.GetHeader("X-Customer-Token")); token != "" {
		return token
	}
	return ""
}
