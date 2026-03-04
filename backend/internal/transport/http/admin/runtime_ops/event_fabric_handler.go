package runtime_ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket/bus"
	"github.com/gin-gonic/gin"
)

type eventFabricTopicItem struct {
	Topic       string   `json:"topic"`
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}

type eventFabricCreateTopicsRequest struct {
	Topic       string                 `json:"topic"`
	Key         string                 `json:"key"`
	Description string                 `json:"description"`
	Actions     []string               `json:"actions"`
	Topics      []eventFabricTopicItem `json:"topics"`
	TenantUUID  string                 `json:"tenant_uuid"`
	TraceID     string                 `json:"trace_id"`
}

type EventFabricHandler struct {
	deps *app.Deps
}

func NewEventFabricHandler(deps *app.Deps) *EventFabricHandler {
	return &EventFabricHandler{deps: deps}
}

func (h *EventFabricHandler) CreateTopics(c *gin.Context) {
	var req eventFabricCreateTopicsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	topics, err := normalizeEventFabricTopics(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth := resolveWSBusGatewayAuth(c, h.deps)
	driver := h.resolveEventTopicDriver(auth)
	tenant, tenantRequired := h.resolveForwardTenant(req.TenantUUID, auth)
	if tenantRequired && tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_uuid is required"})
		return
	}

	if driver == "local" {
		names := make([]string, 0, len(topics))
		for _, topic := range topics {
			names = append(names, topic.Topic)
		}
		bus.DefaultTopicRegistry.Register(names)
		c.JSON(http.StatusAccepted, gin.H{
			"ok":                   true,
			"action":               "create_topics",
			"driver":               driver,
			"topics":               topics,
			"gateway_token_source": auth.Source,
			"tenant_id":            tenant,
			"message":              "event topics created (local registry)",
		})
		return
	}

	if !auth.hasGatewayCredential() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "event fabric gateway credential unavailable", "source": gatewayAuthSource(auth)})
		return
	}

	traceID := strings.TrimSpace(req.TraceID)
	if traceID == "" {
		traceID = strings.TrimSpace(c.GetHeader("X-Request-ID"))
	}
	if err := forwardCreateTopicsToHost(auth, tenant, traceID, topics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"ok":                   true,
		"action":               "create_topics",
		"driver":               driver,
		"topics":               topics,
		"gateway_token_source": auth.Source,
		"tenant_id":            strings.TrimSpace(auth.TenantID),
		"message":              "event topics forwarded to host",
	})
}

func normalizeEventFabricTopics(req eventFabricCreateTopicsRequest) ([]eventFabricTopicItem, error) {
	items := make([]eventFabricTopicItem, 0, len(req.Topics)+1)
	if len(req.Topics) > 0 {
		items = append(items, req.Topics...)
	} else {
		items = append(items, eventFabricTopicItem{
			Topic:       req.Topic,
			Key:         req.Key,
			Description: req.Description,
			Actions:     req.Actions,
		})
	}

	seen := map[string]struct{}{}
	out := make([]eventFabricTopicItem, 0, len(items))
	for _, item := range items {
		topic := strings.TrimSpace(item.Topic)
		if topic == "" {
			topic = strings.TrimSpace(item.Key)
		}
		if topic == "" {
			continue
		}
		if !strings.HasPrefix(topic, "_topic.") {
			return nil, fmt.Errorf("topic %s must start with _topic.", topic)
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		out = append(out, eventFabricTopicItem{
			Topic:       topic,
			Description: strings.TrimSpace(item.Description),
			Actions:     normalizeActions(item.Actions),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("topics is required")
	}
	return out, nil
}

func normalizeActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"publish", "subscribe"}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		v := strings.ToLower(strings.TrimSpace(action))
		if v == "" {
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

func forwardCreateTopicsToHost(auth wsBusGatewayAuthDecision, tenant, traceID string, topics []eventFabricTopicItem) error {
	payload := map[string]any{
		"topics": topics,
	}
	if tenant != "" {
		payload["tenant_uuid"] = tenant
	}
	if traceID != "" {
		payload["trace_id"] = traceID
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request failed")
	}

	endpoint := buildGatewayEndpoint(auth.GatewayBaseURL, auth.GatewayAPIPrefix, "/internal/event-fabric/topics")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build request failed")
	}
	req.Header.Set("Authorization", strings.TrimSpace(auth.Authorization))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if traceID != "" {
		req.Header.Set("X-Request-ID", traceID)
	}

	client := &http.Client{Timeout: auth.GatewayTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("forward failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = fmt.Sprintf("host rejected with status %d", resp.StatusCode)
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func (h *EventFabricHandler) resolveEventTopicDriver(auth wsBusGatewayAuthDecision) string {
	if auth.ProxyEnabled {
		return "framework"
	}
	if h == nil || h.deps == nil || h.deps.Config == nil {
		return "local"
	}
	driver := strings.TrimSpace(h.deps.Config.ResolveEventTopicDriver())
	if driver == "" {
		return "local"
	}
	return driver
}

func (h *EventFabricHandler) resolveForwardTenant(requestTenant string, auth wsBusGatewayAuthDecision) (string, bool) {
	if auth.ProxyEnabled {
		return "", false
	}
	tenant := strings.TrimSpace(auth.TenantID)
	if tenant == "" {
		tenant = strings.TrimSpace(requestTenant)
	}
	return tenant, true
}
