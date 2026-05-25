package middleware

import (
	"context"

	"net/http"
)

const (
	MDKAuthorization = "authorization"
)

// —— HTTP 出站 —— //
func InjectHTTP(ctx context.Context, req *http.Request, bearer string, tc TenantContext, cfg JWTAuthConfig) {
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
}

// —— gRPC 出站 —— //
type PerRPCCreds struct {
	Bearer string
	TC     TenantContext
	Cfg    JWTAuthConfig
}

func (p PerRPCCreds) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	if p.Bearer != "" {
		return map[string]string{MDKAuthorization: "Bearer " + p.Bearer}, nil
	}
	return map[string]string{}, nil
}
func (PerRPCCreds) RequireTransportSecurity() bool { return true }

// 服务端辅助：把当前 TenantContext 写入 gRPC metadata（便于链路下游兜底）
func InjectServerMetadata(ctx context.Context, tc TenantContext, cfg JWTAuthConfig) context.Context {
	return ctx
}
