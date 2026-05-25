package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RBAC 仅做粗粒度权限判定；在 DelegateToPowerX 模式下只校验令牌来源
func RBAC(cfg *authx.RBACConfig, _ authx.ABACClient, _ func(string, string) (bool, map[string]any)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil || !cfg.Enabled || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		// 内部回调放行（仍建议结合网络/签名防护）
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/internal/") || strings.HasPrefix(c.Request.URL.Path, "/api/v1/agent/") {
			c.Next()
			return
		}

		if cfg.DelegateToPowerX {
			if allowPowerXDelegate(c, cfg) {
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Role unauthorized"})
			}
			return
		}

		tc, ok := authx.GetTenantContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Role Authentication required"})
			return
		}
		if authx.IsSuperAdmin(tc.Roles, cfg.SuperAdminRoles) {
			c.Next()
			return
		}

		full := c.FullPath()
		if full == "" {
			full = c.Request.URL.Path
		}
		perm, has := authx.MatchRoute(c.Request.Method, full, cfg.RoutePermissions)
		passRBAC := (!has && !cfg.DefaultDeny) || (has && authx.HasPerm(tc.Permissions, perm))
		if !passRBAC {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":             "Insufficient permissions",
				"required_resource": perm.Resource,
				"required_action":   perm.Action,
			})
			return
		}

		c.Next()
	}
}

func allowPowerXDelegate(c *gin.Context, cfg *authx.RBACConfig) bool {
	if cfg == nil {
		return false
	}
	pxlogger.WithFields(pxlogger.Fields{
		"component":         "http.middleware.rbac",
		"powerx_issuer":     cfg.PowerXIssuer,
		"powerx_audience":   cfg.PowerXAudience,
		"trace_id":          traceIdentifier(c),
		"request_id":        traceIdentifier(c),
		"delegate_to_host":  true,
		"authorization_src": "bearer",
	}).Debug("rbac delegate check")
	if _, ok := authx.GetTenantContext(c); !ok {
		return false
	}
	raw, ok := authx.GetRawBearerToken(c)
	if !ok || strings.TrimSpace(raw) == "" {
		return false
	}
	claims, err := decodeJWTClaims(raw)
	if err != nil {
		return false
	}
	pxlogger.WithFields(pxlogger.Fields{
		"component": "http.middleware.rbac",
		"issuer":    claims["iss"],
		"audience":  claims["aud"],
		"trace_id":  traceIdentifier(c),
	}).Debug("rbac delegate token claims")

	if cfg.PowerXIssuer != "" {
		if iss, _ := claims["iss"].(string); iss != cfg.PowerXIssuer {
			return false
		}
	}
	if cfg.PowerXAudience != "" && !audienceMatches(claims["aud"], cfg.PowerXAudience) {
		return false
	}
	return true
}

func decodeJWTClaims(raw string) (map[string]any, error) {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return nil, errInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func audienceMatches(aud any, expected string) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s == expected {
				return true
			}
		}
	}
	return false
}

var errInvalidToken = errors.New("invalid token")
