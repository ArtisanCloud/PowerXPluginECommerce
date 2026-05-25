package middleware

import (
	"os"
	"strings"
	"time"

	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RequestTrace 输出请求关键信息，辅助排查网关/本地两种模式的差异。
func RequestTrace() gin.HandlerFunc {
	if !traceEnabled() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	mode := requestMode()
	iamMode := iamModeFromEnv()
	return func(c *gin.Context) {
		start := time.Now()

		authMode, authPreview := detectAuth(c)
		userAgent := shorten(c.GetHeader("User-Agent"), 80)
		traceID := traceIdentifier(c)
		tenantCtx, _ := authx.GetTenantContext(c)

		pxlogger.WithFields(pxlogger.Fields{
			"component":   "http.middleware.request_trace",
			"stage":       "begin",
			"mode":        mode,
			"iam_mode":    iamMode,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"auth":        authMode,
			"auth_head":   authPreview,
			"tenant_uuid": tenantCtx.TenantUUID,
			"user_id":     tenantCtx.UserID,
			"trace_id":    traceID,
			"request_id":  traceID,
			"ip":          c.ClientIP(),
			"user_agent":  userAgent,
		}).Debug("request trace begin")

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		if raw, ok := authx.GetRawBearerToken(c); ok && raw != "" {
			authPreview = shorten(raw, 40)
			authMode = "bearer(validated)"
		}

		pxlogger.WithFields(pxlogger.Fields{
			"component":   "http.middleware.request_trace",
			"stage":       "end",
			"mode":        mode,
			"iam_mode":    iamMode,
			"status":      status,
			"latency":     latency.String(),
			"auth":        authMode,
			"auth_head":   authPreview,
			"tenant_uuid": tenantCtx.TenantUUID,
			"user_id":     tenantCtx.UserID,
			"trace_id":    traceID,
			"request_id":  traceID,
		}).Debug("request trace end")
	}
}

func traceEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("POWERX_DEBUG_TRAFFIC")))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	// 默认：PowerX 宿主关闭，独立模式开启
	return os.Getenv("POWERX_PROXY") != "1"
}

func requestMode() string {
	if os.Getenv("POWERX_PROXY") == "1" {
		return "powerx-proxy"
	}
	return "standalone"
}

func detectAuth(c *gin.Context) (mode, preview string) {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		return "bearer", shorten(auth, 40)
	}
	return "none", ""
}

func iamModeFromEnv() string {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("IAM_MODE"))); v != "" {
		if v == "delegated" {
			return "delegated"
		}
		if v == "local" {
			return "local"
		}
	}
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("POWERX_IAM_MODE"))); v != "" {
		if v == "delegated" {
			return "delegated"
		}
		if v == "local" {
			return "local"
		}
	}
	if strings.TrimSpace(os.Getenv("POWERX_PROXY")) == "1" {
		return "delegated"
	}
	return "local"
}

func traceIdentifier(c *gin.Context) string {
	if id := strings.TrimSpace(c.GetHeader("X-Request-ID")); id != "" {
		return id
	}
	if id := strings.TrimSpace(c.GetHeader("Request-ID")); id != "" {
		return id
	}
	if v := strings.TrimSpace(c.GetString("request_id")); v != "" {
		return v
	}
	return ""
}

func shorten(raw string, keep int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len(raw) <= keep {
		return raw
	}
	return raw[:keep] + "..."
}
