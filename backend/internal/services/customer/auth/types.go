package auth

import (
	"context"
	"errors"
	"time"
)

// Context 表示已鉴权客户上下文。
type Context struct {
	TenantUUID string
	CustomerID string
	Roles      []string
	Attributes map[string]string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	RawToken   string
}

// Authenticator 统一客户 token 校验入口。
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*Context, error)
}

// LocalAuthService 定义本地注册/登录能力。
type LocalAuthService interface {
	Authenticator
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
}

// RegisterInput 注册请求负载。
type RegisterInput struct {
	Name       string
	Identifier string
	Password   string
	Email      string
	Phone      string
}

// LoginInput 登录请求负载。
type LoginInput struct {
	Identifier string
	Password   string
}

// AuthResult 返回 token 及客户摘要。
type AuthResult struct {
	Token        string
	ExpiresAt    time.Time
	TenantUUID   string
	CustomerID   string
	CustomerName string
}

var (
	// ErrUnauthorized 表示凭证失效或校验失败。
	ErrUnauthorized = errors.New("customer auth: unauthorized")
	// ErrTokenRequired 表示缺少 token。
	ErrTokenRequired = errors.New("customer auth: token required")
	// ErrIdentifierExists 表示账号重复。
	ErrIdentifierExists = errors.New("customer auth: identifier already exists")
	// ErrInvalidCredentials 表示账号或密码错误。
	ErrInvalidCredentials = errors.New("customer auth: invalid credentials")
	// ErrRegistrationDisabled 当宿主模式关闭本地注册时返回。
	ErrRegistrationDisabled = errors.New("customer auth: registration disabled")
)
