package iam

import (
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func TestLocalIssueTokensUsesCoreXUserMemberClaims(t *testing.T) {
	dir := &LocalDirectory{
		cfg:        &config.Config{},
		issuer:     "powerx-local",
		audience:   "plugin:com.powerx.plugins.ecommerce",
		hmacSecret: []byte("secret"),
		accessTTL:  time.Minute,
		refreshTTL: time.Hour,
	}
	uc := &UserContext{
		TenantUUID:  "00000000-0000-0000-0000-000000000001",
		TenantID:    1,
		UserID:      10,
		MemberID:    20,
		Email:       "USER@EXAMPLE.INVALID",
		Phone:       "0000000000",
		Roles:       []string{"system.admin"},
		Permissions: []string{"*"},
	}

	tokens, err := dir.issueTokens(uc)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims := &authx.PowerXClaims{}
	parsed, err := jwt.ParseWithClaims(tokens.AccessToken, claims, func(token *jwt.Token) (any, error) {
		return []byte("secret"), nil
	}, jwt.WithIssuer("powerx-local"), jwt.WithAudience("plugin:com.powerx.plugins.ecommerce"))
	if err != nil || parsed == nil || !parsed.Valid {
		t.Fatalf("parse issued token: token=%v err=%v", parsed, err)
	}
	if claims.TenantUUID.String() != uc.TenantUUID || claims.TenantID.Int64() != int64(uc.TenantID) {
		t.Fatalf("tenant claims mismatch: %#v", claims)
	}
	if claims.User.UUID == "" || claims.UserID.Int64() != int64(uc.UserID) {
		t.Fatalf("user claims mismatch: %#v", claims)
	}
	if claims.Member.UUID == "" || claims.MemberID.Int64() != int64(uc.MemberID) {
		t.Fatalf("member claims mismatch: %#v", claims)
	}
	if claims.Subject != claims.Member.UUID {
		t.Fatalf("subject should be member uuid, got subject=%q member=%q", claims.Subject, claims.Member.UUID)
	}
	if claims.Scope != "access" {
		t.Fatalf("scope mismatch: %q", claims.Scope)
	}
}
