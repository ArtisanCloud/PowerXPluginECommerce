package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// Handler 暴露 mini-app 客户注册/登录接口（仅在 local 模式下启用）。
type Handler struct {
	mode    config.CustomerAuthMode
	service customerauth.LocalAuthService
}

// NewHandler 构造 handler。
func NewHandler(deps *app.Deps) *Handler {
	if deps == nil {
		return &Handler{mode: config.CustomerAuthModeLocal}
	}
	return &Handler{
		mode:    deps.CustomerAuthMode,
		service: deps.LocalCustomerAuth,
	}
}

// Register 处理 /mini-app/auth/register。
func (h *Handler) Register(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseServiceUnavailable(c, "registration disabled in delegated mode", nil)
		return
	}
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
