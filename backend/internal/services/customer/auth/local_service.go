package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LocalConfig 描述本地 token 签发参数。
type LocalConfig struct {
	JWTSecret []byte
	Issuer    string
	Audience  string
	TTL       time.Duration
}

// LocalService 提供 skeleton 模式下的注册/登录/token 校验。
type LocalService struct {
	db          *gorm.DB
	cfg         LocalConfig
	accounts    *customerrepo.AccountRepository
	customers   *customerrepo.Repository
	tokenLeeway time.Duration
}

// NewLocalService 构造本地客户鉴权服务。
func NewLocalService(db *gorm.DB, cfg LocalConfig) (*LocalService, error) {
	if db == nil {
		return nil, errors.New("customer auth: db dependency required")
	}
	if len(cfg.JWTSecret) == 0 {
		return nil, errors.New("customer auth: jwt secret required")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 2 * time.Hour
	}
	return &LocalService{
		db:          db,
		cfg:         cfg,
		accounts:    customerrepo.NewAccountRepository(db),
		customers:   customerrepo.NewRepository(db),
		tokenLeeway: time.Minute,
	}, nil
}

// Register 创建账号并返回 token。
func (s *LocalService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	if s == nil {
		return nil, ErrRegistrationDisabled
	}
	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}
	identifier := normalizeIdentifier(input.Identifier)
	if existing, err := s.accounts.FindByIdentifier(ctx, identifier); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, ErrIdentifierExists
	}
	customerEntity := &customermodel.Customer{
		CustomerID:     utils.NewUUID(),
		Name:           strings.TrimSpace(input.Name),
		Type:           customerTypeFromIdentifier(identifier),
		Email:          optionalEmail(identifier, input.Email),
		Phone:          optionalPhone(identifier, input.Phone),
		Source:         "mini-app",
		Status:         "active",
		MembershipTier: "standard",
	}
	if _, err := s.customers.CreateCustomer(ctx, customerEntity); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(input.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	account := &customermodel.CustomerAccount{
		CustomerID:   customerEntity.CustomerID,
		Identifier:   identifier,
		PasswordHash: string(hash),
		Status:       "active",
	}
	if _, err := s.accounts.CreateAccount(ctx, account); err != nil {
		return nil, err
	}
	return s.issueToken(ctx, customerEntity.CustomerID, customerEntity.Name)
}

// Login 校验账号密码并返回 token。
func (s *LocalService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	if err := validateLoginInput(input); err != nil {
		return nil, err
	}
	account, err := s.accounts.FindByIdentifier(ctx, normalizeIdentifier(input.Identifier))
	if err != nil {
		return nil, err
	}
	if account == nil || account.Status != "active" {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(strings.TrimSpace(input.Password))); err != nil {
		return nil, ErrInvalidCredentials
	}
	customerEntity, err := s.customers.FindByCustomerID(ctx, account.CustomerID)
	if err != nil {
		return nil, err
	}
	if customerEntity == nil {
		return nil, ErrInvalidCredentials
	}
	if err := s.accounts.UpdateAccount(ctx, account, map[string]any{
		"last_login_at": time.Now(),
	}); err != nil {
		return nil, err
	}
	result, err := s.issueToken(ctx, account.CustomerID, customerEntity.Name)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Authenticate 验证 token 并返回上下文。
func (s *LocalService) Authenticate(ctx context.Context, token string) (*Context, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrTokenRequired
	}
	customClaims := &customerClaims{}
	parsed, err := jwt.ParseWithClaims(strings.TrimSpace(token), customClaims, func(t *jwt.Token) (any, error) {
		return s.cfg.JWTSecret, nil
	}, jwt.WithAudience(s.cfg.Audience), jwt.WithIssuer(s.cfg.Issuer), jwt.WithLeeway(s.tokenLeeway))
	if err != nil || parsed == nil || !parsed.Valid {
		return nil, ErrUnauthorized
	}
	if customClaims.TenantUUID == "" || customClaims.CustomerID == "" {
		return nil, ErrUnauthorized
	}
	claims := customClaims.RegisteredClaims
	return &Context{
		TenantUUID: customClaims.TenantUUID,
		CustomerID: customClaims.CustomerID,
		Roles:      customClaims.Roles,
		IssuedAt:   claims.IssuedAt.Time,
		ExpiresAt:  claims.ExpiresAt.Time,
		RawToken:   token,
	}, nil
}

func (s *LocalService) issueToken(ctx context.Context, customerID, customerName string) (*AuthResult, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	exp := now.Add(s.cfg.TTL)
	claims := customerClaims{
		TenantUUID: tenantUUID,
		CustomerID: customerID,
		Roles:      []string{"customer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Audience:  jwt.ClaimStrings{s.cfg.Audience},
			Subject:   customerID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now.Add(-s.tokenLeeway)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		Token:        signed,
		ExpiresAt:    exp,
		TenantUUID:   tenantUUID,
		CustomerID:   customerID,
		CustomerName: customerName,
	}, nil
}

func validateRegisterInput(input RegisterInput) error {
	identifier := strings.TrimSpace(input.Identifier)
	password := strings.TrimSpace(input.Password)
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	if identifier == "" {
		return errors.New("identifier is required")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func validateLoginInput(input LoginInput) error {
	if strings.TrimSpace(input.Identifier) == "" || strings.TrimSpace(input.Password) == "" {
		return errors.New("identifier and password are required")
	}
	return nil
}

func normalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

func optionalEmail(identifier, fallback string) string {
	id := strings.TrimSpace(identifier)
	if strings.Contains(id, "@") {
		return id
	}
	return strings.TrimSpace(fallback)
}

func optionalPhone(identifier, fallback string) string {
	id := strings.TrimSpace(identifier)
	if phoneCandidate(id) {
		return id
	}
	return strings.TrimSpace(fallback)
}

func phoneCandidate(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(value) >= 6
}

func customerTypeFromIdentifier(identifier string) string {
	if strings.Contains(identifier, "@") {
		return "individual"
	}
	return "individual"
}

type customerClaims struct {
	TenantUUID string   `json:"tid"`
	CustomerID string   `json:"cid"`
	Roles      []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}
