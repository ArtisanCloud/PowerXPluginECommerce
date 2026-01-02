package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

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
					if uid, ok := m["uid"].(float64); ok {
						tc.UserID = int64(uid)
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
		log.Printf("[PLUGIN-JWT-AUTH] JWTAuth failed. cfg{Issuer=%s, AcceptAudiences=%v, Optional=%v}. RawAuth=%s",
			cfg.Issuer, cfg.AcceptAudiences, cfg.Optional, shorten(rawAuth, 40),
		)
		if os.Getenv("POWERX_DEBUG_TRAFFIC") == "1" && strings.HasPrefix(strings.ToLower(rawAuth), "bearer ") {
			tok := strings.TrimSpace(rawAuth[len("Bearer "):])
			if m, err := decodeJWTClaims(tok); err == nil {
				log.Printf("[PLUGIN-JWT-AUTH][TOKEN] iss=%v aud=%v sub=%v iat=%v nbf=%v exp=%v",
					m["iss"], m["aud"], m["sub"], m["iat"], m["nbf"], m["exp"])
			}
			if _, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
				return []byte(cfg.HMACSecret), nil
			}, jwt.WithAudience(cfg.AcceptAudiences...), jwt.WithIssuer(cfg.Issuer)); err != nil {
				log.Printf("[PLUGIN-JWT-AUTH][VERIFY] %v", err)
			} else {
				log.Printf("[PLUGIN-JWT-AUTH][VERIFY] ok")
			}
			log.Printf("[PLUGIN-JWT-AUTH][CFG] issuer=%s audiences=%v secret.len=%d",
				cfg.Issuer, cfg.AcceptAudiences, len(cfg.HMACSecret))
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "jwt Unauthorized"})
	}
}
