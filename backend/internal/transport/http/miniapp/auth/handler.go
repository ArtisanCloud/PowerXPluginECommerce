package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	minibase "github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram/base/request"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	pxmodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	adminpayments "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Handler 暴露 mini-app 客户注册/登录接口（仅在 local 模式下启用）。
type Handler struct {
	mode       config.CustomerAuthMode
	service    customerauth.LocalAuthService
	deps       *app.Deps
	accounts   *customerrepo.AccountRepository
	customers  *customerrepo.Repository
	identities *customerrepo.IdentityRepository
}

// NewHandler 构造 handler。
func NewHandler(deps *app.Deps) *Handler {
	if deps == nil {
		return &Handler{mode: config.CustomerAuthModeLocal}
	}
	return &Handler{
		mode:       deps.CustomerAuthMode,
		service:    deps.LocalCustomerAuth,
		deps:       deps,
		accounts:   customerrepo.NewAccountRepository(deps.DB),
		customers:  customerrepo.NewRepository(deps.DB),
		identities: customerrepo.NewIdentityRepository(deps.DB),
	}
}

// Register 处理 /mini-app/auth/register。
func (h *Handler) Register(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "registration disabled in delegated mode", nil)
		return
	}
	attachTenantToRequest(c)
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	result, err := h.service.Register(c.Request.Context(), req.toInput())
	if err != nil {
		if errors.Is(err, customerauth.ErrIdentifierExists) {
			contracts.ResponseError(c, http.StatusConflict, "customer.identifier_exists", err.Error())
			return
		}
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, newAuthResponse(result))
}

// Login 处理 /mini-app/auth/login。
func (h *Handler) Login(c *gin.Context) {
	if h.mode == config.CustomerAuthModeDelegate && h.service == nil {
		contracts.ResponseServiceUnavailable(c, "login delegated to host", nil)
		return
	}
	attachTenantToRequest(c)
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	result, err := h.service.Login(c.Request.Context(), req.toInput())
	if err != nil {
		if errors.Is(err, customerauth.ErrInvalidCredentials) {
			contracts.ResponseUnauthorized(c, "invalid credentials")
			return
		}
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, newAuthResponse(result))
}

// WechatLogin 处理 /mini-app/auth/wechat/login。
func (h *Handler) WechatLogin(c *gin.Context) {
	if h.mode == config.CustomerAuthModeDelegate && h.service == nil {
		contracts.ResponseServiceUnavailable(c, "login delegated to host", nil)
		return
	}
	if h.deps == nil || h.deps.DB == nil {
		contracts.ResponseServiceUnavailable(c, "database unavailable", nil)
		return
	}
	attachTenantToRequest(c)
	var req wechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	providerRow, err := h.loadWechatProvider(c, req.ProviderID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	app, err := h.buildMiniProgramAppFromProvider(providerRow)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	session, err := app.Auth.Session(c.Request.Context(), strings.TrimSpace(req.Code))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	if session == nil || session.ErrCode != 0 || strings.TrimSpace(session.OpenID) == "" {
		contracts.ResponseBadRequest(c, "wechat session invalid")
		return
	}
	appID := strings.TrimSpace(app.GetConfig().GetString("app_id", ""))
	identifier := "wechat:" + strings.TrimSpace(session.OpenID)
	provider := "wechat"
	subject := strings.TrimSpace(session.OpenID)
	unionID := strings.TrimSpace(session.UnionID)
	identity, err := h.identities.FindByProviderAppSubject(c.Request.Context(), provider, appID, subject)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	if identity == nil && appID != "" {
		if legacy, err := h.identities.FindByProviderSubjectLegacy(c.Request.Context(), provider, subject); err == nil && legacy != nil {
			_ = h.identities.UpdateIdentity(c.Request.Context(), legacy, map[string]any{"app_id": appID})
			identity = legacy
		}
	}

	var customerEntity *customer.Customer
	if identity != nil {
		customerEntity, err = h.customers.FindByCustomerID(c.Request.Context(), identity.CustomerID)
		if err != nil {
			contracts.ResponseBadRequest(c, err.Error())
			return
		}
		if customerEntity == nil {
			contracts.ResponseBadRequest(c, "customer not found")
			return
		}
		identityUpdates := map[string]any{}
		if unionID != "" && strings.TrimSpace(identity.UnionID) == "" {
			identityUpdates["union_id"] = unionID
		}
		if nickname := strings.TrimSpace(req.Nickname); nickname != "" || strings.TrimSpace(req.AvatarURL) != "" {
			identityMeta, _ := json.Marshal(map[string]any{
				"app_id":     appID,
				"nickname":   strings.TrimSpace(req.Nickname),
				"avatar_url": strings.TrimSpace(req.AvatarURL),
			})
			identityUpdates["metadata"] = datatypes.JSON(identityMeta)
		}
		if len(identityUpdates) > 0 {
			_ = h.identities.UpdateIdentity(c.Request.Context(), identity, identityUpdates)
		}
		if customerEntity != nil {
			nickname := strings.TrimSpace(req.Nickname)
			avatarURL := strings.TrimSpace(req.AvatarURL)
			if nickname != "" || avatarURL != "" {
				updatedMeta := map[string]any{}
				if len(customerEntity.Metadata) > 0 {
					_ = json.Unmarshal(customerEntity.Metadata, &updatedMeta)
				}
				if nickname != "" {
					customerEntity.Name = nickname
					updatedMeta["nickname"] = nickname
				}
				if avatarURL != "" {
					updatedMeta["avatar_url"] = avatarURL
				}
				if len(updatedMeta) > 0 {
					metaRaw, _ := json.Marshal(updatedMeta)
					customerEntity.Metadata = datatypes.JSON(metaRaw)
				}
				_, _ = h.customers.SaveCustomer(c.Request.Context(), customerEntity)
			}
		}
	} else {
		// Fallback: migrate existing wechat account record to identity binding.
		if h.accounts != nil {
			if account, err := h.accounts.FindByIdentifier(c.Request.Context(), identifier); err == nil && account != nil {
				customerEntity, err = h.customers.FindByCustomerID(c.Request.Context(), account.CustomerID)
				if err != nil {
					contracts.ResponseBadRequest(c, err.Error())
					return
				}
				if customerEntity == nil {
					contracts.ResponseBadRequest(c, "customer not found")
					return
				}
			}
		}
		if customerEntity == nil {
			metaRaw, _ := json.Marshal(map[string]any{
				"wechat_appid":   appID,
				"wechat_openid":  subject,
				"wechat_unionid": unionID,
				"avatar_url":     strings.TrimSpace(req.AvatarURL),
				"nickname":       strings.TrimSpace(req.Nickname),
			})
			customerEntity = &customer.Customer{
				CustomerID:     utils.NewUUID(),
				Name:           fallbackName(req.Nickname),
				Type:           "individual",
				Source:         "mini-app",
				Status:         "active",
				MembershipTier: "standard",
				Metadata:       datatypes.JSON(metaRaw),
			}
			if _, err := h.customers.CreateCustomer(c.Request.Context(), customerEntity); err != nil {
				contracts.ResponseBadRequest(c, err.Error())
				return
			}
		}
		identityMeta, _ := json.Marshal(map[string]any{
			"app_id":     appID,
			"nickname":   strings.TrimSpace(req.Nickname),
			"avatar_url": strings.TrimSpace(req.AvatarURL),
		})
		if _, err := h.identities.UpsertIdentity(c.Request.Context(), &customer.CustomerIdentity{
			CustomerID: customerEntity.CustomerID,
			Provider:   provider,
			AppID:      appID,
			Subject:    subject,
			UnionID:    unionID,
			Metadata:   datatypes.JSON(identityMeta),
		}); err != nil {
			contracts.ResponseBadRequest(c, err.Error())
			return
		}
	}

	localSvc, ok := h.service.(*customerauth.LocalService)
	if !ok || localSvc == nil {
		contracts.ResponseServiceUnavailable(c, "local auth service unavailable", nil)
		return
	}
	result, err := localSvc.IssueTokenForCustomer(c.Request.Context(), customerEntity.CustomerID, customerEntity.Name)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	resp := wechatLoginResponse{
		authResponse: newAuthResponse(result),
		OpenID:       strings.TrimSpace(session.OpenID),
		UnionID:      strings.TrimSpace(session.UnionID),
	}
	contracts.ResponseSuccess(c, resp)
}

// WechatPhoneNumber 处理 /mini-app/auth/wechat/phone。
func (h *Handler) WechatPhoneNumber(c *gin.Context) {
	attachTenantToRequest(c)
	var req wechatPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	providerRow, err := h.loadWechatProvider(c, req.ProviderID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	app, err := h.buildMiniProgramAppFromProvider(providerRow)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	resp, err := app.PhoneNumber.GetUserPhoneNumber(c.Request.Context(), strings.TrimSpace(req.Code))
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

// WechatDecryptData 处理 /mini-app/auth/wechat/decrypt。
func (h *Handler) WechatDecryptData(c *gin.Context) {
	attachTenantToRequest(c)
	var req wechatDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	providerRow, err := h.loadWechatProvider(c, req.ProviderID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	app, err := h.buildMiniProgramAppFromProvider(providerRow)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	data, cryptErr := app.Encryptor.DecryptData(
		strings.TrimSpace(req.EncryptedData),
		strings.TrimSpace(req.SessionKey),
		strings.TrimSpace(req.IV),
	)
	if cryptErr != nil {
		contracts.ResponseBadRequest(c, cryptErr.ErrMsg)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"data": string(data)})
}

// WechatCheckEncryptedData 处理 /mini-app/auth/wechat/check-encrypted。
func (h *Handler) WechatCheckEncryptedData(c *gin.Context) {
	attachTenantToRequest(c)
	var req wechatCheckEncryptedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	providerRow, err := h.loadWechatProvider(c, req.ProviderID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	app, err := h.buildMiniProgramAppFromProvider(providerRow)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	hash := strings.TrimSpace(req.Hash)
	if hash == "" && strings.TrimSpace(req.EncryptedData) != "" {
		sum := sha256.Sum256([]byte(req.EncryptedData))
		hash = hex.EncodeToString(sum[:])
	}
	if hash == "" {
		contracts.ResponseBadRequest(c, "hash is required")
		return
	}
	resp, err := app.Base.CheckEncryptedData(c.Request.Context(), hash)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

// WechatPaidUnionID 处理 /mini-app/auth/wechat/paid-unionid。
func (h *Handler) WechatPaidUnionID(c *gin.Context) {
	attachTenantToRequest(c)
	var req wechatPaidUnionIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	providerRow, err := h.loadWechatProvider(c, req.ProviderID)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	app, err := h.buildMiniProgramAppFromProvider(providerRow)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	params := &minibase.RequestGetPaidUnionID{
		OpenID:        strings.TrimSpace(req.OpenID),
		TransactionID: strings.TrimSpace(req.TransactionID),
		MchID:         strings.TrimSpace(req.MchID),
		OutTradeNo:    strings.TrimSpace(req.OutTradeNo),
	}
	if params.OpenID == "" {
		contracts.ResponseBadRequest(c, "openid is required")
		return
	}
	resp, err := app.Base.GetPaidUnionID(c.Request.Context(), params)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, resp)
}

type registerRequest struct {
	Name       string `json:"name" binding:"required"`
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

func (r registerRequest) toInput() customerauth.RegisterInput {
	return customerauth.RegisterInput{
		Name:       r.Name,
		Identifier: r.Identifier,
		Password:   r.Password,
		Email:      r.Email,
		Phone:      r.Phone,
	}
}

type loginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (r loginRequest) toInput() customerauth.LoginInput {
	return customerauth.LoginInput{
		Identifier: r.Identifier,
		Password:   r.Password,
	}
}

type authResponse struct {
	Token        string `json:"token"`
	ExpiresAt    string `json:"expiresAt"`
	CustomerID   string `json:"customerId"`
	TenantUUID   string `json:"tenantUuid"`
	CustomerName string `json:"customerName"`
}

type wechatLoginRequest struct {
	ProviderID string `json:"providerId" binding:"required"`
	Code       string `json:"code" binding:"required"`
	Nickname   string `json:"nickname"`
	AvatarURL  string `json:"avatarUrl"`
}

type wechatPhoneRequest struct {
	ProviderID string `json:"providerId" binding:"required"`
	Code       string `json:"code" binding:"required"`
}

type wechatDecryptRequest struct {
	ProviderID    string `json:"providerId" binding:"required"`
	EncryptedData string `json:"encryptedData" binding:"required"`
	SessionKey    string `json:"sessionKey" binding:"required"`
	IV            string `json:"iv" binding:"required"`
}

type wechatCheckEncryptedRequest struct {
	ProviderID    string `json:"providerId" binding:"required"`
	EncryptedData string `json:"encryptedData"`
	Hash          string `json:"hash"`
}

type wechatPaidUnionIDRequest struct {
	ProviderID    string `json:"providerId" binding:"required"`
	OpenID        string `json:"openid" binding:"required"`
	TransactionID string `json:"transactionId"`
	MchID         string `json:"mchId"`
	OutTradeNo    string `json:"outTradeNo"`
}

type wechatLoginResponse struct {
	authResponse
	OpenID  string `json:"openid"`
	UnionID string `json:"unionid"`
}

func newAuthResponse(result *customerauth.AuthResult) authResponse {
	resp := authResponse{}
	if result == nil {
		return resp
	}
	resp.Token = result.Token
	resp.CustomerID = result.CustomerID
	resp.CustomerName = result.CustomerName
	resp.TenantUUID = result.TenantUUID
	resp.ExpiresAt = formatTime(result.ExpiresAt)
	return resp
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func attachTenantToRequest(c *gin.Context) {
	if c == nil {
		return
	}
	if tc, ok := authx.GetTenantContext(c); ok && tc.TenantUUID != "" {
		ctx := authx.ContextWithTenantUUID(c.Request.Context(), tc.TenantUUID)
		if ctx != nil {
			c.Request = c.Request.WithContext(ctx)
		}
	}
}

func (h *Handler) buildMiniProgramAppFromProvider(provider *pxmodels.PaymentProvider) (*miniProgram.MiniProgram, error) {
	if h == nil || h.deps == nil || h.deps.Config == nil || provider == nil {
		return nil, errors.New("wechat provider config missing")
	}
	creds, err := adminpayments.DecryptProviderCredentials(h.deps.Config, provider.TenantUuid, provider.ID, provider.Credentials)
	if err != nil {
		return nil, err
	}
	appID := strings.TrimSpace(provider.AppID)
	if appID == "" {
		appID = credentialString(creds, "appId", "app_id")
	}
	secret := credentialString(creds, "appSecret", "app_secret", "secret")
	if appID == "" || secret == "" {
		return nil, errors.New("wechat miniapp credentials missing")
	}
	httpDebug, hasHttpDebug := credentialBool(creds, "httpDebug", "http_debug")
	debug, hasDebug := credentialBool(creds, "debug")
	if h.deps.Config.Server != nil && h.deps.Config.Server.DevMode {
		if !hasHttpDebug {
			httpDebug = true
		}
		if !hasDebug {
			debug = true
		}
	}
	return miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     appID,
		Secret:    secret,
		HttpDebug: httpDebug,
		Debug:     debug,
		Log: miniProgram.Log{
			Level:  "debug",
			ENV:    "develop",
			Stdout: true,
		},
	})
}

func fallbackName(nickname string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		return nickname
	}
	return "微信用户"
}

func (h *Handler) loadWechatProvider(c *gin.Context, providerID string) (*pxmodels.PaymentProvider, error) {
	if h == nil || h.deps == nil || h.deps.DB == nil {
		return nil, errors.New("payment provider repository unavailable")
	}
	idRaw := strings.TrimSpace(providerID)
	if idRaw == "" {
		return nil, errors.New("providerId is required")
	}
	idNum, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil || idNum == 0 {
		return nil, errors.New("invalid providerId")
	}
	tenantUUID, err := authx.RequireTenantUUID(c.Request.Context())
	if err != nil {
		return nil, err
	}
	var row pxmodels.PaymentProvider
	if err := h.deps.DB.WithContext(c.Request.Context()).
		Where("tenant_uuid = ? AND id = ? AND status = ?", tenantUUID, idNum, "active").
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wechat provider not found")
		}
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(row.ProviderType), "wechat") {
		return nil, errors.New("wechat provider required")
	}
	return &row, nil
}

func credentialString(payload map[string]interface{}, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			default:
				b, err := json.Marshal(t)
				if err == nil {
					value := strings.Trim(string(b), "\"")
					if strings.TrimSpace(value) != "" {
						return strings.TrimSpace(value)
					}
				}
			}
		}
	}
	return ""
}

func keysOfCredentials(payload map[string]interface{}) []string {
	if payload == nil {
		return nil
	}
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func credentialBool(payload map[string]interface{}, keys ...string) (bool, bool) {
	if payload == nil {
		return false, false
	}
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case bool:
				return t, true
			case string:
				if strings.EqualFold(strings.TrimSpace(t), "true") {
					return true, true
				}
				if strings.EqualFold(strings.TrimSpace(t), "false") {
					return false, true
				}
			case float64:
				return t != 0, true
			case int:
				return t != 0, true
			default:
				if b, err := json.Marshal(t); err == nil {
					s := strings.Trim(string(b), "\"")
					if strings.EqualFold(strings.TrimSpace(s), "true") {
						return true, true
					}
					if strings.EqualFold(strings.TrimSpace(s), "false") {
						return false, true
					}
				}
			}
		}
	}
	return false, false
}
