package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseFromHeadersReadsCoreXUserAndMemberClaims(t *testing.T) {
	now := time.Now()
	claims := PowerXClaims{
		TenantUUID: TenantClaim("tenant-uuid"),
		TenantID:   FlexibleInt(1),
		User:       IdentityClaim{UUID: "user-uuid"},
		UserID:     FlexibleInt(10),
		Member:     IdentityClaim{UUID: "member-uuid"},
		MemberID:   FlexibleInt(20),
		Email:      "USER@EXAMPLE.INVALID",
		Phone:      "0000000000",
		Scope:      "access",
		Roles:      []string{"system.admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "powerx-local",
			Subject:   "member-uuid",
			Audience:  jwt.ClaimStrings{"plugin:com.powerx.plugins.ecommerce"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	tc, _, ok := ParseFromHeaders(func(key string) string {
		if key == "Authorization" {
			return "Bearer " + signed
		}
		return ""
	}, JWTAuthConfig{
		Issuer:          "powerx-local",
		AcceptAudiences: []string{"plugin:com.powerx.plugins.ecommerce"},
		HMACSecret:      "secret",
	})
	if !ok {
		t.Fatal("expected token to parse")
	}
	if tc.TenantUUID != "tenant-uuid" || tc.TenantID != 1 {
		t.Fatalf("tenant claims mismatch: %#v", tc)
	}
	if tc.UserUUID != "user-uuid" || tc.UserID != 10 {
		t.Fatalf("user claims mismatch: %#v", tc)
	}
	if tc.MemberUUID != "member-uuid" || tc.MemberID != 20 {
		t.Fatalf("member claims mismatch: %#v", tc)
	}
	if tc.Email != "user@example.invalid" || tc.Phone != "0000000000" {
		t.Fatalf("contact claims mismatch: %#v", tc)
	}
}

func TestParseFromHeadersKeepsLegacyNumericUID(t *testing.T) {
	now := time.Now()
	claims := jwt.MapClaims{
		"tid": "tenant-uuid",
		"uid": 10,
		"iss": "powerx-local",
		"aud": "plugin:com.powerx.plugins.ecommerce",
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(time.Minute).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	tc, _, ok := ParseFromHeaders(func(key string) string {
		if key == "Authorization" {
			return "Bearer " + signed
		}
		return ""
	}, JWTAuthConfig{
		Issuer:          "powerx-local",
		AcceptAudiences: []string{"plugin:com.powerx.plugins.ecommerce"},
		HMACSecret:      "secret",
	})
	if !ok {
		t.Fatal("expected legacy token to parse")
	}
	if tc.UserID != 10 || tc.UserUUID != "" {
		t.Fatalf("legacy uid mismatch: %#v", tc)
	}
}
