package public

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	middleware "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/authproxy"
	iamservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
)

// authProxy captures the delegated client surface the handler needs.
type authProxy interface {
	Login(ctx context.Context, req iamservice.LoginRequest) (*iamservice.AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (*iamservice.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	MeContext(ctx context.Context, accessToken string) (*authproxy.MeContext, error)
}

// AuthHandler exposes /api/v1/auth public endpoints.
type AuthHandler struct {
	mode  iamservice.IAMMode
	proxy authProxy
	local iamservice.IAMDirectory
}

// NewAuthHandler builds a handler for the given IAM mode.
func NewAuthHandler(deps *app.Deps) *AuthHandler {
	if deps == nil {
		return &AuthHandler{}
	}
	return &AuthHandler{mode: deps.IAMMode, proxy: deps.AuthProxy, local: deps.IAMDirectory}
}

// RegisterAuthRoutes wires /auth routes beneath the API prefix.
func RegisterAuthRoutes(group *gin.RouterGroup, deps *app.Deps) {
	if group == nil || deps == nil {
		return
	}
	handler := NewAuthHandler(deps)
	adminAuth := group.Group("/admin/user/auth")
	adminAuth.Use(middleware.RequestTrace())
	adminAuth.POST("/login", handler.Login)
	adminAuth.POST("/refresh", handler.Refresh)
	adminAuth.POST("/logout", handler.Logout)
	adminAuth.GET("/me/context", handler.MeContext)
}

// Login proxies login requests to PowerX Core.
func (h *AuthHandler) Login(c *gin.Context) {
	switch h.mode {
	case iamservice.IAMModeDelegated:
		if !h.ensureDelegated(c) {
			return
		}
		h.handleDelegatedLogin(c)
	case iamservice.IAMModeLocal:
		h.handleLocalLogin(c)
	default:
		contracts.ResponseServiceUnavailable(c, "当前 IAM 模式未启用", nil)
	}
}

// Refresh exchanges refresh_token for a new access token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	switch h.mode {
	case iamservice.IAMModeDelegated:
		if !h.ensureDelegated(c) {
			return
		}
		h.handleDelegatedRefresh(c)
	case iamservice.IAMModeLocal:
		h.handleLocalRefresh(c)
	default:
		contracts.ResponseServiceUnavailable(c, "当前 IAM 模式未启用", nil)
	}
}

// Logout revokes the current refresh token upstream.
func (h *AuthHandler) Logout(c *gin.Context) {
	switch h.mode {
	case iamservice.IAMModeDelegated:
		if !h.ensureDelegated(c) {
			return
		}
		h.handleDelegatedLogout(c)
	case iamservice.IAMModeLocal:
		h.handleLocalLogout(c)
	default:
		contracts.ResponseServiceUnavailable(c, "当前 IAM 模式未启用", nil)
	}
}

// MeContext fetches the active user context from PowerX Core.
func (h *AuthHandler) MeContext(c *gin.Context) {
	switch h.mode {
	case iamservice.IAMModeDelegated:
		if !h.ensureDelegated(c) {
			return
		}
		h.handleDelegatedMeContext(c)
	case iamservice.IAMModeLocal:
		h.handleLocalMeContext(c)
	default:
		contracts.ResponseServiceUnavailable(c, "当前 IAM 模式未启用", nil)
	}
}

func (h *AuthHandler) ensureDelegated(c *gin.Context) bool {
	if h.mode != iamservice.IAMModeDelegated {
		contracts.ResponseServiceUnavailable(c, "当前路由仅支持 Delegated 模式", nil)
		return false
	}
	if h.proxy == nil {
		contracts.ResponseServiceUnavailable(c, "宿主认证未配置", nil)
		return false
	}
	return true
}

func (h *AuthHandler) handleProxyErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, iamservice.ErrAuthUnavailable):
		authmetrics.RecordDelegateError("unavailable")
		contracts.ResponseServiceUnavailable(c, "宿主认证不可用，请稍后重试", nil)
	case errors.Is(err, iamservice.ErrUnauthorized):
		authmetrics.RecordDelegateError("unauthorized")
		contracts.ResponseUnauthorized(c, "认证失败，请重新登录")
	default:
		var perr *authproxy.ProxyError
		if errors.As(err, &perr) {
			authmetrics.RecordDelegateError("proxy")
			contracts.ResponseError(c, perr.Status, contracts.ErrCodeInternalError, perr.Message)
		} else {
			authmetrics.RecordDelegateError("other")
			contracts.ResponseInternalError(c, err)
		}
	}
}

func (h *AuthHandler) handleLocalErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, iamservice.ErrUnauthorized):
		contracts.ResponseUnauthorized(c, "认证失败，请重新登录")
	case errors.Is(err, iamservice.ErrInvalidArguments):
		contracts.ResponseBadRequest(c, "请求参数无效")
	case errors.Is(err, iamservice.ErrAuthUnavailable):
		contracts.ResponseServiceUnavailable(c, "本地认证暂不可用", nil)
	default:
		contracts.ResponseInternalError(c, err)
	}
}

func (h *AuthHandler) handleDelegatedLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "参数错误: "+err.Error())
		return
	}
	tokens, err := h.proxy.Login(c.Request.Context(), iamservice.LoginRequest{
		Tenant:     req.Tenant,
		Identifier: req.Identifier,
		Password:   req.Password,
		Remember:   req.Remember,
	})
	if err != nil {
		authmetrics.RecordLogin(h.modeLabel(), "failure")
		h.handleProxyErr(c, err)
		return
	}
	authmetrics.RecordLogin(h.modeLabel(), "success")
	contracts.ResponseSuccess(c, mapTokens(tokens))
}

func (h *AuthHandler) handleDelegatedRefresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		contracts.ResponseBadRequest(c, "refresh_token 必填")
		return
	}
	tokens, err := h.proxy.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		authmetrics.RecordRefresh(h.modeLabel(), "failure")
		h.handleProxyErr(c, err)
		return
	}
	authmetrics.RecordRefresh(h.modeLabel(), "success")
	contracts.ResponseSuccess(c, mapTokens(tokens))
}

func (h *AuthHandler) handleDelegatedLogout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		contracts.ResponseBadRequest(c, "refresh_token 必填")
		return
	}
	if err := h.proxy.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleProxyErr(c, err)
		return
	}
	authmetrics.RecordLogout(h.modeLabel())
	contracts.ResponseSuccess(c, gin.H{"ok": true})
}

func (h *AuthHandler) handleDelegatedMeContext(c *gin.Context) {
	token := extractBearer(c.GetHeader("Authorization"))
	if token == "" {
		contracts.ResponseUnauthorized(c, "缺少 Authorization Bearer token")
		return
	}
	ctx, err := h.proxy.MeContext(c.Request.Context(), token)
	if err != nil {
		h.handleProxyErr(c, err)
		return
	}
	contracts.ResponseSuccess(c, mapDelegatedContext(ctx))
}

func (h *AuthHandler) handleLocalLogin(c *gin.Context) {
	if h.local == nil {
		contracts.ResponseServiceUnavailable(c, "本地 IAM 未初始化", nil)
		return
	}
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "参数错误: "+err.Error())
		return
	}
	tokens, _, err := h.local.Login(c.Request.Context(), iamservice.LoginRequest{
		Tenant:     req.Tenant,
		Identifier: req.Identifier,
		Password:   req.Password,
		Remember:   req.Remember,
	})
	if err != nil {
		authmetrics.RecordLogin(h.modeLabel(), "failure")
		h.handleLocalErr(c, err)
		return
	}
	authmetrics.RecordLogin(h.modeLabel(), "success")
	contracts.ResponseSuccess(c, mapTokens(tokens))
}

func (h *AuthHandler) handleLocalRefresh(c *gin.Context) {
	if h.local == nil {
		contracts.ResponseServiceUnavailable(c, "本地 IAM 未初始化", nil)
		return
	}
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		contracts.ResponseBadRequest(c, "refresh_token 必填")
		return
	}
	tokens, err := h.local.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		authmetrics.RecordRefresh(h.modeLabel(), "failure")
		h.handleLocalErr(c, err)
		return
	}
	authmetrics.RecordRefresh(h.modeLabel(), "success")
	contracts.ResponseSuccess(c, mapTokens(tokens))
}

func (h *AuthHandler) handleLocalLogout(c *gin.Context) {
	if h.local == nil {
		contracts.ResponseServiceUnavailable(c, "本地 IAM 未初始化", nil)
		return
	}
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		contracts.ResponseBadRequest(c, "refresh_token 必填")
		return
	}
	if err := h.local.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleLocalErr(c, err)
		return
	}
	authmetrics.RecordLogout(h.modeLabel())
	contracts.ResponseSuccess(c, gin.H{"ok": true})
}

func (h *AuthHandler) handleLocalMeContext(c *gin.Context) {
	if h.local == nil {
		contracts.ResponseServiceUnavailable(c, "本地 IAM 未初始化", nil)
		return
	}
	token := extractBearer(c.GetHeader("Authorization"))
	if token == "" {
		contracts.ResponseUnauthorized(c, "缺少 Authorization Bearer token")
		return
	}
	decoder, ok := h.local.(interface {
		UserContextFromToken(context.Context, string) (*iamservice.UserContext, error)
	})
	if !ok {
		contracts.ResponseServiceUnavailable(c, "本地 IAM 不支持上下文解析", nil)
		return
	}
	uc, err := decoder.UserContextFromToken(c.Request.Context(), token)
	if err != nil {
		h.handleLocalErr(c, err)
		return
	}
	contracts.ResponseSuccess(c, mapUserContext(uc))
}

func mapUserContext(uc *iamservice.UserContext) gin.H {
	if uc == nil {
		return gin.H{}
	}
	roles := normalizeStringList(uc.Roles)
	permissions := normalizeStringList(uc.Permissions)
	isRoot := hasAnyRole(roles, "system.admin", "superadmin", "root")
	canManageTemplates := isRoot || hasAnyRole(roles, "admin", "tenant.admin", "template.admin") || hasPermission(permissions, "base.templates.manage")
	tenantUUID := strings.TrimSpace(uc.TenantUUID)
	tenant := gin.H{
		"uuid": tenantUUID,
		"key":  uc.TenantKey,
		"name": uc.TenantName,
	}
	if legacyRaw := strings.TrimSpace(uc.TenantUuid); legacyRaw != "" {
		if legacyID, err := strconv.ParseUint(legacyRaw, 10, 64); err == nil && legacyID > 0 {
			tenant["legacy_id"] = legacyID
		}
	}
	return gin.H{
		"is_root":             isRoot,
		"current_tenant_uuid": tenantUUID,
		"current_member_id":   uc.MemberID,
		"tenant":              tenant,
		"user": gin.H{
			"id":           uc.UserID,
			"username":     uc.Username,
			"email":        uc.Email,
			"display_name": uc.DisplayName,
		},
		"members": []gin.H{
			{
				"tenant_uuid": tenantUUID,
				"tenant_name": uc.TenantName,
				"member_id":   uc.MemberID,
				"is_admin":    canManageTemplates,
			},
		},
		"roles":       roles,
		"permissions": permissions,
		"capabilities": gin.H{
			"templates": gin.H{
				"can_create": canManageTemplates,
				"can_update": canManageTemplates,
				"can_delete": canManageTemplates,
			},
		},
	}
}

func mapDelegatedContext(mc *authproxy.MeContext) gin.H {
	if mc == nil {
		return gin.H{}
	}
	roles := normalizeStringList(mc.Roles)
	permissions := normalizeStringList(mc.Permissions)
	currentTenantUUID := strings.TrimSpace(mc.CurrentTenantUUID)
	currentMemberAdmin := false
	for _, member := range mc.Members {
		if strings.TrimSpace(member.TenantUUID) == currentTenantUUID && member.IsAdmin {
			currentMemberAdmin = true
			break
		}
	}
	canManageTemplates := mc.IsRoot || currentMemberAdmin || hasPermission(permissions, "base.templates.manage")
	canCreate := canManageTemplates
	canUpdate := canManageTemplates
	canDelete := canManageTemplates
	if mc.Capabilities != nil && mc.Capabilities.Templates != nil {
		canCreate = mc.Capabilities.Templates.CanCreate || canManageTemplates
		canUpdate = mc.Capabilities.Templates.CanUpdate || canManageTemplates
		canDelete = mc.Capabilities.Templates.CanDelete || canManageTemplates
	}

	return gin.H{
		"is_root":             mc.IsRoot,
		"current_tenant_uuid": currentTenantUUID,
		"current_member_id":   mc.CurrentMemberID,
		"user":                mc.User,
		"members":             mc.Members,
		"roles":               roles,
		"permissions":         permissions,
		"capabilities": gin.H{
			"templates": gin.H{
				"can_create": canCreate,
				"can_update": canUpdate,
				"can_delete": canDelete,
			},
		},
	}
}

func normalizeStringList(input []string) []string {
	if len(input) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		val := strings.TrimSpace(item)
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		out = append(out, val)
	}
	return out
}

func hasAnyRole(roles []string, candidates ...string) bool {
	if len(roles) == 0 || len(candidates) == 0 {
		return false
	}
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		roleSet[strings.ToLower(strings.TrimSpace(role))] = struct{}{}
	}
	for _, candidate := range candidates {
		key := strings.ToLower(strings.TrimSpace(candidate))
		if key == "" {
			continue
		}
		if _, ok := roleSet[key]; ok {
			return true
		}
	}
	return false
}

func hasPermission(permissions []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" || len(permissions) == 0 {
		return false
	}
	targetLower := strings.ToLower(target)
	for _, permission := range permissions {
		key := strings.ToLower(strings.TrimSpace(permission))
		if key == targetLower {
			return true
		}
	}
	return false
}

func mapTokens(tokens *iamservice.AuthTokens) gin.H {
	if tokens == nil {
		return gin.H{}
	}
	expiresAt := tokens.ExpiresAt.UTC()
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}
	return gin.H{
		"token_type":    tokens.TokenType,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"expires_at":    expiresAt.UnixMilli(),
		"scope":         tokens.Scope,
	}
}

func extractBearer(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[len("Bearer "):])
	}
	return ""
}

func (h *AuthHandler) modeLabel() string {
	if h == nil {
		return ""
	}
	return string(h.mode)
}

type loginRequest struct {
	Tenant     string `json:"tenant"`
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Remember   bool   `json:"remember"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
