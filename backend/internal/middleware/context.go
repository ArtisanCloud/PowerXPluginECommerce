package middleware

// internal/middleware/context.go

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TenantContext struct {
	TenantUUID    string   `json:"tenant_uuid"`
	TenantID      int64    `json:"tenant_id,omitempty"`
	UserID        int64    `json:"user_id"`
	UserUUID      string   `json:"user_uuid,omitempty"`
	MemberID      int64    `json:"member_id,omitempty"`
	MemberUUID    string   `json:"member_uuid,omitempty"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	IsRoot        bool     `json:"is_root,omitempty"`
	Platforms     []string `json:"platforms,omitempty"`
	Roles         []string `json:"roles"`
	Permissions   []string `json:"permissions"`
	PolicyVersion string   `json:"policy_version"`
}

type ActorContext struct {
	TenantUUID string `json:"tenant_uuid,omitempty"`
	TenantID   int64  `json:"tenant_id,omitempty"`
	UserUUID   string `json:"user_uuid,omitempty"`
	UserID     int64  `json:"user_id,omitempty"`
	MemberUUID string `json:"member_uuid,omitempty"`
	MemberID   int64  `json:"member_id,omitempty"`
	Email      string `json:"email,omitempty"`
}

func (tc TenantContext) Actor() ActorContext {
	return ActorContext{
		TenantUUID: strings.TrimSpace(tc.TenantUUID),
		TenantID:   tc.TenantID,
		UserUUID:   strings.TrimSpace(tc.UserUUID),
		UserID:     tc.UserID,
		MemberUUID: strings.TrimSpace(tc.MemberUUID),
		MemberID:   tc.MemberID,
		Email:      strings.ToLower(strings.TrimSpace(tc.Email)),
	}
}

func (a ActorContext) ID() string {
	if v := strings.TrimSpace(a.MemberUUID); v != "" {
		return "member:" + v
	}
	if a.MemberID > 0 {
		return "member_id:" + strconv.FormatInt(a.MemberID, 10)
	}
	if v := strings.TrimSpace(a.UserUUID); v != "" {
		return "user:" + v
	}
	if a.UserID > 0 {
		return "user_id:" + strconv.FormatInt(a.UserID, 10)
	}
	if v := strings.TrimSpace(a.TenantUUID); v != "" {
		return "tenant:" + v
	}
	return "system"
}

func (a ActorContext) Fields() map[string]any {
	fields := map[string]any{}
	if v := strings.TrimSpace(a.TenantUUID); v != "" {
		fields["tenant_uuid"] = v
	}
	if a.TenantID > 0 {
		fields["tenant_id"] = a.TenantID
	}
	if v := strings.TrimSpace(a.UserUUID); v != "" {
		fields["user_uuid"] = v
	}
	if a.UserID > 0 {
		fields["user_id"] = a.UserID
	}
	if v := strings.TrimSpace(a.MemberUUID); v != "" {
		fields["member_uuid"] = v
	}
	if a.MemberID > 0 {
		fields["member_id"] = a.MemberID
	}
	if v := strings.TrimSpace(a.Email); v != "" {
		fields["email"] = v
	}
	return fields
}

func ActorFromContext(ctx context.Context) ActorContext {
	if tc, ok := TenantContextFromContext(ctx); ok {
		return tc.Actor()
	}
	if tenantUUID, ok := TenantUUIDFromContext(ctx); ok {
		return ActorContext{TenantUUID: tenantUUID}
	}
	return ActorContext{}
}

func ActorIDFromContext(ctx context.Context) string {
	return ActorFromContext(ctx).ID()
}

func ActorFromGin(c *gin.Context) (ActorContext, bool) {
	if c == nil {
		return ActorContext{}, false
	}
	if tc, ok := GetTenantContext(c); ok {
		return tc.Actor(), true
	}
	if c.Request != nil {
		actor := ActorFromContext(c.Request.Context())
		if strings.TrimSpace(actor.ID()) != "system" {
			return actor, true
		}
	}
	return ActorContext{}, false
}

func RequireActorFromGin(c *gin.Context) (ActorContext, bool) {
	actor, ok := ActorFromGin(c)
	if !ok || actor.ID() == "system" {
		return ActorContext{}, false
	}
	return actor, true
}

func ActorIDFromGin(c *gin.Context) string {
	if actor, ok := ActorFromGin(c); ok {
		return actor.ID()
	}
	return "system"
}

// CustomerContext 存放 mini-app 客户信息。
type CustomerContext struct {
	TenantUUID string   `json:"tenant_uuid"`
	CustomerID string   `json:"customer_id"`
	Roles      []string `json:"roles"`
}

const (
	ctxKeyTenant   = "tenant_ctx"
	ctxKeyToken    = "raw_bearer_token"
	ctxKeyCustomer = "customer_ctx"
)

type stdCtxKey string

const (
	stdTenantCtxKey stdCtxKey = stdCtxKey(ctxKeyTenant)
	stdRawBearerKey stdCtxKey = stdCtxKey(ctxKeyToken)
	stdCustomerKey  stdCtxKey = stdCtxKey(ctxKeyCustomer)
)

type tenantUUIDContextKey struct{}
type customerContextKey struct{}
type requestIDContextKey struct{}

var ctxKeyTenantUUID = tenantUUIDContextKey{}
var ctxKeyCustomerCtx = customerContextKey{}
var ctxKeyRequestID = requestIDContextKey{}

var ErrTenantMissing = errors.New("tenant context missing")

func SetTenantContext(c *gin.Context, tc TenantContext) { c.Set(ctxKeyTenant, tc) }
func GetTenantContext(c *gin.Context) (TenantContext, bool) {
	v, ok := c.Get(ctxKeyTenant)
	if !ok || v == nil {
		return TenantContext{}, false
	}
	tc, ok := v.(TenantContext)
	return tc, ok
}
func SetRawBearerToken(c *gin.Context, token string) {
	if token != "" {
		c.Set(ctxKeyToken, token)
	}
}
func GetRawBearerToken(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxKeyToken)
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// ContextWithTenantContext stores TenantContext into a standard context.
// 为了兼容旧代码，同时写入 string key 与 typed key。
func ContextWithTenantContext(ctx context.Context, tc TenantContext) context.Context {
	if ctx == nil {
		return nil
	}
	ctx = context.WithValue(ctx, stdTenantCtxKey, tc)
	return ctx
}

// TenantContextFromContext extracts TenantContext from a standard context.
// 为了兼容旧代码，同时读取 string key 与 typed key。
func TenantContextFromContext(ctx context.Context) (TenantContext, bool) {
	if ctx == nil {
		return TenantContext{}, false
	}
	if v := ctx.Value(stdTenantCtxKey); v != nil {
		if tc, ok := v.(TenantContext); ok {
			return tc, true
		}
	}
	if v := ctx.Value(ctxKeyTenant); v != nil {
		if tc, ok := v.(TenantContext); ok {
			return tc, true
		}
	}
	return TenantContext{}, false
}

// ContextWithRawBearerToken stores raw bearer token into a standard context.
// 为了兼容旧代码，同时写入 string key 与 typed key。
func ContextWithRawBearerToken(ctx context.Context, token string) context.Context {
	if ctx == nil {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, stdRawBearerKey, token)
	return ctx
}

// RawBearerTokenFromContext extracts raw bearer token from a standard context.
// 为了兼容旧代码，同时读取 string key 与 typed key。
func RawBearerTokenFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v := ctx.Value(stdRawBearerKey); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), true
		}
	}
	if v := ctx.Value(ctxKeyToken); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), true
		}
	}
	return "", false
}

// SetCustomerContext stores customer info on gin context.
func SetCustomerContext(c *gin.Context, cc CustomerContext) {
	if c == nil {
		return
	}
	c.Set(ctxKeyCustomer, cc)
	if cc.CustomerID != "" {
		ctx := context.WithValue(c.Request.Context(), ctxKeyCustomerCtx, cc)
		c.Request = c.Request.WithContext(ctx)
	}
}

// GetCustomerContext retrieves customer context.
func GetCustomerContext(c *gin.Context) (CustomerContext, bool) {
	if v, ok := c.Get(ctxKeyCustomer); ok {
		if cc, ok := v.(CustomerContext); ok {
			return cc, true
		}
	}
	if c.Request != nil {
		if ctx := c.Request.Context(); ctx != nil {
			if v := ctx.Value(ctxKeyCustomerCtx); v != nil {
				if cc, ok := v.(CustomerContext); ok {
					return cc, true
				}
			}
		}
	}
	return CustomerContext{}, false
}

// ContextWithTenantUUID stores tenant UUID into a standard context.
func ContextWithTenantUUID(ctx context.Context, tenantUUID string) context.Context {
	if ctx == nil || strings.TrimSpace(tenantUUID) == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyTenantUUID, strings.TrimSpace(tenantUUID))
}

// TenantUUIDFromContext extracts tenant UUID from a standard context.
func TenantUUIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v := ctx.Value(ctxKeyTenantUUID); v != nil {
		switch id := v.(type) {
		case string:
			if strings.TrimSpace(id) != "" {
				return strings.TrimSpace(id), true
			}
		case uint64:
			if id > 0 {
				return strconv.FormatUint(id, 10), true
			}
		case int64:
			if id > 0 {
				return strconv.FormatInt(id, 10), true
			}
		case int:
			if id > 0 {
				return strconv.Itoa(id), true
			}
		}
	}
	return "", false
}

// RequireTenantUUID retrieves tenant UUID from context or returns ErrTenantMissing.
func RequireTenantUUID(ctx context.Context) (string, error) {
	if tenantUUID, ok := TenantUUIDFromContext(ctx); ok && tenantUUID != "" {
		return tenantUUID, nil
	}
	return "", ErrTenantMissing
}

// ContextWithRequestID stores request id into a standard context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil || strings.TrimSpace(requestID) == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyRequestID, strings.TrimSpace(requestID))
}

// RequestIDFromContext extracts request id from a standard context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v := ctx.Value(ctxKeyRequestID); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), true
		}
	}
	// Backward compatibility with gin key usage.
	if v := ctx.Value("request_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), true
		}
	}
	return "", false
}

// CustomerFromContext returns customer context if exists.
func CustomerFromContext(ctx context.Context) (CustomerContext, bool) {
	if ctx == nil {
		return CustomerContext{}, false
	}
	if v := ctx.Value(ctxKeyCustomerCtx); v != nil {
		if cc, ok := v.(CustomerContext); ok {
			return cc, true
		}
	}
	return CustomerContext{}, false
}

// Deprecated compatibility helpers —— convert numeric IDs into UUID strings if possible.
func ContextWithTenantUuid(ctx context.Context, tenantID uint64) context.Context {
	if tenantID == 0 {
		return ctx
	}
	return ContextWithTenantUUID(ctx, strconv.FormatUint(tenantID, 10))
}

func TenantUuidFromContext(ctx context.Context) (uint64, bool) {
	if uuidVal, ok := TenantUUIDFromContext(ctx); ok && uuidVal != "" {
		if num, err := strconv.ParseUint(uuidVal, 10, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

func RequireTenantUuid(ctx context.Context) (uint64, error) {
	if id, ok := TenantUuidFromContext(ctx); ok && id > 0 {
		return id, nil
	}
	return 0, ErrTenantMissing
}
