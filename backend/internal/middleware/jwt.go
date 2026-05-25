package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthConfig struct {
	Issuer             string   `yaml:"issuer" json:"issuer"`
	AcceptAudiences    []string `yaml:"accept_audiences" json:"accept_audiences"`
	HMACSecret         string   `yaml:"hmac_secret" json:"hmac_secret"`
	ClockSkewSeconds   int      `yaml:"clock_skew_seconds" json:"clock_skew_seconds"`
	Optional           bool     `yaml:"optional" json:"optional"`
	AllowSignedContext bool     `yaml:"allow_signed_context" json:"allow_signed_context"`
	ContextHMACSecret  string   `yaml:"context_hmac_secret" json:"context_hmac_secret"`
	MaxCtxAgeSeconds   int64    `yaml:"max_ctx_age_seconds" json:"max_ctx_age_seconds"`
}

type PowerXClaims struct {
	TenantUUID    TenantClaim `json:"tid"`
	UserID        int64       `json:"uid"`
	Roles         []string    `json:"roles"`
	Permissions   []string    `json:"perms"`
	PolicyVersion string      `json:"policy_version"`
	jwt.RegisteredClaims
}

type TenantClaim string

func (t *TenantClaim) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*t = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*t = TenantClaim(strings.TrimSpace(s))
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*t = TenantClaim(strings.TrimSpace(num.String()))
	return nil
}

func (t TenantClaim) String() string {
	return string(t)
}

func ParseFromHeaders(h func(string) string, cfg JWTAuthConfig) (tc TenantContext, rawBearer string, ok bool) {
	// 1) Authorization: Bearer
	authz := h("Authorization")
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		raw := strings.TrimSpace(authz[7:])
		if raw != "" && cfg.HMACSecret != "" {
			if t, err := parseHS256(raw, cfg); err == nil {
				return t, raw, true
			}
		}
	}
	return TenantContext{}, "", false
}

func parseHS256(raw string, cfg JWTAuthConfig) (TenantContext, error) {
	leeway := time.Duration(cfg.ClockSkewSeconds)
	if leeway <= 0 {
		leeway = 60
	}
	claims := &PowerXClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected sign method")
		}
		return []byte(cfg.HMACSecret), nil
	}, jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.AcceptAudiences...), jwt.WithLeeway(leeway*time.Second))
	if err != nil || token == nil || !token.Valid {
		return TenantContext{}, errors.New("invalid token")
	}
	return TenantContext{
		TenantUUID: strings.TrimSpace(claims.TenantUUID.String()), UserID: claims.UserID, Roles: claims.Roles,
		Permissions: claims.Permissions, PolicyVersion: claims.PolicyVersion,
	}, nil
}
