package runtime_ops

import (
	"net/http"
	"strings"

	frameworkwsbus "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/runtime/wsbus"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket/bus"
	"github.com/gin-gonic/gin"
)

type wsBusGrantRequest struct {
	Topics     []string `json:"topics"`
	Actions    []string `json:"actions"`
	TenantUUID string   `json:"tenant_uuid"`
	TraceID    string   `json:"trace_id"`
}

type wsBusPublishRequest struct {
	Topic      string `json:"topic"`
	Payload    any    `json:"payload"`
	TenantUUID string `json:"tenant_uuid"`
	TraceID    string `json:"trace_id"`
}

type WSBusHandler struct {
	deps *app.Deps
}

func NewWSBusHandler(deps *app.Deps) *WSBusHandler {
	return &WSBusHandler{deps: deps}
}

func (h *WSBusHandler) Grant(c *gin.Context) {
	var req wsBusGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	auth := resolveWSBusGatewayAuth(c, h.deps)
	topics := normalizeTopicsForRegister(req.Topics)
	if len(topics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topics is required"})
		return
	}
	tenant, tenantRequired := h.resolveForwardTenant(req.TenantUUID, auth)
	if tenantRequired && tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_uuid is required"})
		return
	}
	driver := h.resolveEventTopicDriver(auth)

	h.logResolvedAuth("grant."+driver, strings.Join(topics, ","), auth)
	for _, topic := range topics {
		if !bus.DefaultTopicRegistry.Exists(topic) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "topic not found", "topic": topic, "hint": "create topic before grant"})
			return
		}
		bus.DefaultACLRegistry.Grant(tenant, topic, req.Actions)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"ok":                   true,
		"action":               "grant",
		"topics":               topics,
		"actions":              normalizeGrantActions(req.Actions),
		"driver":               driver,
		"gateway_token_source": auth.Source,
		"tenant_id":            tenant,
		"proxied":              auth.ProxyEnabled,
		"iam_mode":             auth.IAMMode,
		"message":              "ws bus grant applied (local acl)",
	})
}

func (h *WSBusHandler) Publish(c *gin.Context) {
	var req wsBusPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	req.Topic = strings.TrimSpace(req.Topic)
	if req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic is required"})
		return
	}

	auth := resolveWSBusGatewayAuth(c, h.deps)
	tenant, tenantRequired := h.resolveForwardTenant(req.TenantUUID, auth)
	if tenantRequired && tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_uuid is required"})
		return
	}
	driver := h.resolveEventTopicDriver(auth)

	h.logResolvedAuth("publish."+driver, req.Topic, auth)
	traceID := strings.TrimSpace(req.TraceID)
	if traceID == "" {
		traceID = strings.TrimSpace(c.GetHeader("X-Request-ID"))
	}
	if !bus.DefaultTopicRegistry.Exists(req.Topic) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic not found", "topic": req.Topic, "hint": "create topic before publish"})
		return
	}
	if !bus.DefaultACLRegistry.Allowed(tenant, req.Topic, "publish") {
		c.JSON(http.StatusForbidden, gin.H{"error": "publish not granted", "topic": req.Topic, "tenant_uuid": tenant})
		return
	}
	result := bus.DefaultPublisher.Publish(c.Request.Context(), req.Topic, req.Payload, frameworkwsbus.PublishOptions{
		TenantUUID: tenant,
		TraceID:    traceID,
	})
	if !result.OK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.ErrorMessage, "code": result.ErrorCode})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"ok":                   true,
		"action":               "publish",
		"topic":                req.Topic,
		"payload":              req.Payload,
		"driver":               driver,
		"gateway_token_source": auth.Source,
		"tenant_id":            tenant,
		"proxied":              auth.ProxyEnabled,
		"iam_mode":             auth.IAMMode,
		"message":              "ws bus publish accepted (local driver)",
	})
}

func (h *WSBusHandler) resolveForwardTenant(requestTenant string, auth wsBusGatewayAuthDecision) (string, bool) {
	if auth.ProxyEnabled {
		// Proxy 模式由宿主根据 token/api key 解析租户，插件不透传 tenant_uuid。
		return "", false
	}
	tenant := strings.TrimSpace(auth.TenantID)
	if tenant == "" {
		tenant = strings.TrimSpace(requestTenant)
	}
	return tenant, true
}

func (h *WSBusHandler) resolveEventTopicDriver(auth wsBusGatewayAuthDecision) string {
	if h == nil || h.deps == nil || h.deps.Config == nil {
		return "local"
	}
	driver := strings.TrimSpace(h.deps.Config.ResolveEventTopicDriver())
	if driver == "" {
		return "local"
	}
	return driver
}

func normalizeTopicsForRegister(topics []string) []string {
	if len(topics) == 0 {
		return nil
	}
	out := make([]string, 0, len(topics))
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		t := strings.TrimSpace(topic)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func (h *WSBusHandler) logResolvedAuth(action string, topic string, auth wsBusGatewayAuthDecision) {
	logger.WithFields(logger.Fields{
		"component":           "runtime_ws_bus",
		"action":              action,
		"topic":               topic,
		"gateway_auth_source": auth.Source,
		"gateway_auth_scheme": auth.GatewayAuthScheme,
		"tenant_id":           auth.TenantID,
		"powerx_proxy":        auth.ProxyEnabled,
		"iam_mode":            auth.IAMMode,
	}).Info("WS bus gateway auth resolved")
}

func normalizeGrantActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"publish", "subscribe"}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		v := strings.ToLower(strings.TrimSpace(action))
		if v != "publish" && v != "subscribe" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return []string{"publish", "subscribe"}
	}
	return out
}
