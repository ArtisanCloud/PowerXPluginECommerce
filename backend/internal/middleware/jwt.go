package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
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
	TenantUUID    TenantClaim   `json:"tid"`
	TenantID      FlexibleInt   `json:"tid_n,omitempty"`
	User          IdentityClaim `json:"uid,omitempty"`
	UserID        FlexibleInt   `json:"uid_n,omitempty"`
	Member        IdentityClaim `json:"mid,omitempty"`
	MemberID      FlexibleInt   `json:"mid_n,omitempty"`
	Email         string        `json:"email,omitempty"`
	Phone         string        `json:"phone,omitempty"`
	IsRoot        bool          `json:"is_root,omitempty"`
	Platforms     []string      `json:"plats,omitempty"`
	Scope         string        `json:"scope,omitempty"`
	Roles         []string      `json:"roles"`
	Permissions   []string      `json:"perms"`
	PolicyVersion string        `json:"policy_version"`
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

type FlexibleInt int64

func (n *FlexibleInt) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*n = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*n = 0
			return nil
		}
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			*n = 0
			return nil
		}
		*n = FlexibleInt(parsed)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	parsed, err := strconv.ParseInt(num.String(), 10, 64)
	if err != nil {
		return err
	}
	*n = FlexibleInt(parsed)
	return nil
}

func (n FlexibleInt) Int64() int64 {
	return int64(n)
}

type IdentityClaim struct {
	UUID string
	ID   FlexibleInt
}

func (c *IdentityClaim) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*c = IdentityClaim{}
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*c = IdentityClaim{}
			return nil
		}
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.ID = FlexibleInt(parsed)
			c.UUID = ""
			return nil
		}
		c.UUID = s
		c.ID = 0
		return nil
	}
	var id FlexibleInt
	if err := id.UnmarshalJSON(data); err != nil {
		return err
	}
	c.ID = id
	c.UUID = ""
	return nil
}

func (c IdentityClaim) MarshalJSON() ([]byte, error) {
	if value := strings.TrimSpace(c.UUID); value != "" {
		return json.Marshal(value)
	}
	if c.ID.Int64() > 0 {
		return json.Marshal(c.ID.Int64())
	}
	return []byte("null"), nil
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
		TenantUUID:    strings.TrimSpace(claims.TenantUUID.String()),
		TenantID:      claims.TenantID.Int64(),
		UserID:        firstInt64(claims.UserID.Int64(), claims.User.ID.Int64()),
		UserUUID:      strings.TrimSpace(claims.User.UUID),
		MemberID:      firstInt64(claims.MemberID.Int64(), claims.Member.ID.Int64()),
		MemberUUID:    strings.TrimSpace(claims.Member.UUID),
		Email:         strings.ToLower(strings.TrimSpace(claims.Email)),
		Phone:         strings.TrimSpace(claims.Phone),
		IsRoot:        claims.IsRoot,
		Platforms:     claims.Platforms,
		Roles:         claims.Roles,
		Permissions:   claims.Permissions,
		PolicyVersion: claims.PolicyVersion,
	}, nil
}

func firstInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
