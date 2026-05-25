package middleware

import (
	"net/http"
	"os"
	"strings"

	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(cfg authx.JWTAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第一通道：标准解析（Bearer -> HS256 claims / 或签名上下文）
		if tc, bearer, ok := authx.ParseFromHeaders(c.GetHeader, cfg); ok {
			authx.SetTenantContext(c, tc)
			authx.SetRawBearerToken(c, bearer)
			c.Next()
			return
		}

		// 调试或容错通道：Bearer 存在且能用 HS256 验签，就从 claims 提取最小上下文并放行
		rawAuth := c.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(rawAuth), "bearer ") && cfg.HMACSecret != "" {
			tok := strings.TrimSpace(rawAuth[len("Bearer "):])

			// 二次验证（与调试打印一致）：Issuer/Audience + HS256
			if _, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
				return []byte(cfg.HMACSecret), nil
			}, jwt.WithAudience(cfg.AcceptAudiences...), jwt.WithIssuer(cfg.Issuer)); err == nil {
				// ✅ 验签成功——尽可能从 token claims 补齐 tid/uid/roles/perms，避免 RBAC 退化为匿名
				tc := authx.TenantContext{}
				if m, err := decodeJWTClaims(tok); err == nil {
					if tid, ok := m["tid"].(string); ok {
						tc.TenantUUID = strings.TrimSpace(tid)
					}
					if tidN, ok := m["tid_n"].(float64); ok {
						tc.TenantID = int64(tidN)
					}
					if uidN, ok := m["uid_n"].(float64); ok {
						tc.UserID = int64(uidN)
					} else if uid, ok := m["uid"].(float64); ok {
						tc.UserID = int64(uid)
					}
					if uid, ok := m["uid"].(string); ok {
						tc.UserUUID = strings.TrimSpace(uid)
					}
					if midN, ok := m["mid_n"].(float64); ok {
						tc.MemberID = int64(midN)
					} else if mid, ok := m["mid"].(float64); ok {
						tc.MemberID = int64(mid)
					}
					if mid, ok := m["mid"].(string); ok {
						tc.MemberUUID = strings.TrimSpace(mid)
					}
					if email, ok := m["email"].(string); ok {
						tc.Email = strings.ToLower(strings.TrimSpace(email))
					}
					if phone, ok := m["phone"].(string); ok {
						tc.Phone = strings.TrimSpace(phone)
					}
					if isRoot, ok := m["is_root"].(bool); ok {
						tc.IsRoot = isRoot
					}
					if plats, ok := m["plats"].([]any); ok {
						for _, p := range plats {
							if s, ok := p.(string); ok && strings.TrimSpace(s) != "" {
								tc.Platforms = append(tc.Platforms, strings.TrimSpace(s))
							}
						}
					}
					if roles, ok := m["roles"].([]any); ok {
						for _, r := range roles {
							if s, ok := r.(string); ok && strings.TrimSpace(s) != "" {
								tc.Roles = append(tc.Roles, strings.TrimSpace(s))
							}
						}
					}
					if perms, ok := m["perms"].([]any); ok {
						for _, p := range perms {
							if s, ok := p.(string); ok && strings.TrimSpace(s) != "" {
								tc.Permissions = append(tc.Permissions, strings.TrimSpace(s))
							}
						}
					}
					if pv, ok := m["policy_version"].(string); ok {
						tc.PolicyVersion = strings.TrimSpace(pv)
					}
				}
				authx.SetTenantContext(c, tc)
				authx.SetRawBearerToken(c, tok)
				c.Next()
				return
			}
		}

		// 走到这里说明双通道都失败
		if cfg.Optional {
			c.Next()
			return
		}

		// 原有的调试日志 + 401
		pxlogger.WithFields(pxlogger.Fields{
			"component":        "http.middleware.jwt_auth",
			"status":           "unauthorized",
			"reason":           "jwt_auth_failed",
			"issuer":           cfg.Issuer,
			"accept_audiences": cfg.AcceptAudiences,
			"optional":         cfg.Optional,
			"auth_head":        shorten(rawAuth, 40),
			"trace_id":         traceIdentifier(c),
			"request_id":       traceIdentifier(c),
		}).Warn("jwt auth failed")
		if os.Getenv("POWERX_DEBUG_TRAFFIC") == "1" && strings.HasPrefix(strings.ToLower(rawAuth), "bearer ") {
			tok := strings.TrimSpace(rawAuth[len("Bearer "):])
			if m, err := decodeJWTClaims(tok); err == nil {
				pxlogger.WithFields(pxlogger.Fields{
					"component": "http.middleware.jwt_auth",
					"issuer":    m["iss"],
					"audience":  m["aud"],
					"subject":   m["sub"],
					"iat":       m["iat"],
					"nbf":       m["nbf"],
					"exp":       m["exp"],
					"trace_id":  traceIdentifier(c),
				}).Debug("jwt token debug claims")
			}
			if _, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
				return []byte(cfg.HMACSecret), nil
			}, jwt.WithAudience(cfg.AcceptAudiences...), jwt.WithIssuer(cfg.Issuer)); err != nil {
				pxlogger.WithError(err).WithField("component", "http.middleware.jwt_auth").Debug("jwt verify debug failed")
			} else {
				pxlogger.WithField("component", "http.middleware.jwt_auth").Debug("jwt verify debug ok")
			}
			pxlogger.WithFields(pxlogger.Fields{
				"component":         "http.middleware.jwt_auth",
				"issuer":            cfg.Issuer,
				"accept_audiences":  cfg.AcceptAudiences,
				"hmac_secret_bytes": len(cfg.HMACSecret),
			}).Debug("jwt auth debug config")
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "jwt Unauthorized"})
	}
}
