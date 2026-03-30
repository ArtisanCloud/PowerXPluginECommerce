package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	integrationService "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/integration"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type capabilityInvokeRequest struct {
	CapabilityID      string         `json:"capabilityId"`
	Action            string         `json:"action"`
	PreferredProtocol string         `json:"preferredProtocol"`
	Payload           map[string]any `json:"payload"`
}

// Handler 提供 integration HTTP API 的入口。
type Handler struct {
	deps     *app.Deps
	dispatch *integrationService.DispatchService
	logger   *logrus.Entry
}

// NewHandler 构造新的 Handler。
func NewHandler(deps *app.Deps) *Handler {
	var logger *logrus.Entry
	if deps != nil {
		logger = deps.RuntimeLogger(deps.Ctx, "integration_http", nil)
	}
	h := &Handler{
		deps:   deps,
		logger: logger,
	}
	h.dispatch = h.buildDispatchService()
	return h
}

func (h *Handler) buildDispatchService() *integrationService.DispatchService {
	if h.deps == nil {
		return nil
	}

	logger := h.logger
	if logger == nil {
		logger = logrus.WithField("component", "integration_http")
	}
	service := integrationService.BuildDispatchService(h.deps, logger)
	if service == nil {
		return nil
	}
	return service
}

// ListGrantMatrix 返回当前 GrantMatrix 视图。
func (h *Handler) ListGrantMatrix(c *gin.Context) {
	respondPlaceholder(c, http.StatusOK, "grant matrix listing not implemented")
}

// SubmitGrantMatrix 接收数据库覆盖项。
func (h *Handler) SubmitGrantMatrix(c *gin.Context) {
	respondPlaceholder(c, http.StatusAccepted, "grant matrix override submission pending approval workflow")
}

// CreateSubscription 注册 webhook 订阅。
func (h *Handler) CreateSubscription(c *gin.Context) {
	respondPlaceholder(c, http.StatusCreated, "webhook subscription endpoint not implemented")
}

// ListSubscriptions 查询 webhook 订阅。
func (h *Handler) ListSubscriptions(c *gin.Context) {
	respondPlaceholder(c, http.StatusOK, "webhook subscription list not implemented")
}

// ReplayDLQ 触发 DLQ 补发。
func (h *Handler) ReplayDLQ(c *gin.Context) {
	respondPlaceholder(c, http.StatusAccepted, "webhook DLQ replay not implemented")
}

// CreateSecret 注册外部凭证。
func (h *Handler) CreateSecret(c *gin.Context) {
	respondPlaceholder(c, http.StatusCreated, "secret lifecycle endpoint not implemented")
}

// RotateSecret 触发凭证轮换。
func (h *Handler) RotateSecret(c *gin.Context) {
	respondPlaceholder(c, http.StatusAccepted, "secret rotation workflow not implemented")
}

// InvokeCapability 兼容 Capability Lab 调试入口，转发到网关能力调用接口。
func (h *Handler) InvokeCapability(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request body"}})
		return
	}

	base := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	if base == "" {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "failed to invoke capability via gateway",
				"details": gin.H{"error": "gateway base not configured"},
			},
		})
		return
	}
	base = ensureAPIV1(base)
	primaryEndpoint := strings.TrimRight(base, "/") + "/tenant/invocations"
	legacyEndpoint := strings.TrimRight(base, "/") + "/integration/capabilities/invoke"

	requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	tenantUUID := strings.TrimSpace(c.GetHeader("X-Tenant-UUID"))
	if tenantUUID == "" {
		tenantUUID = firstNonEmptyEnv("PX_TENANT_UUID", "POWERX_TENANT_UUID")
	}
	mockModule := strings.TrimSpace(c.GetHeader("X-PX-Use-Mock"))

	primaryBody, err := buildTenantInvocationPayload(rawBody, tenantUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "invalid capability payload",
				"details": gin.H{"error": err.Error()},
			},
		})
		return
	}

	scheme := resolveGatewayAuthScheme()
	token := firstNonEmptyEnv("PX_TOOL_TOKEN", "PX_PLUGIN_TOOL_TOKEN", "POWERX_AUTH_TOKEN")
	apiKey := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY"))

	status, headers, raw, err := invokeGatewayEndpoint(
		c,
		primaryEndpoint,
		primaryBody,
		scheme,
		token,
		apiKey,
		requestID,
		tenantUUID,
		mockModule,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "failed to invoke capability via gateway",
				"details": gin.H{"error": err.Error(), "endpoint": primaryEndpoint},
			},
		})
		return
	}
	// 兼容旧网关：tenant/invocations 不存在时回退旧路径
	if status == http.StatusNotFound {
		status, headers, raw, err = invokeGatewayEndpoint(
			c,
			legacyEndpoint,
			rawBody,
			scheme,
			token,
			apiKey,
			requestID,
			tenantUUID,
			mockModule,
		)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"message": "failed to invoke capability via gateway",
					"details": gin.H{"error": err.Error(), "endpoint": legacyEndpoint},
				},
			})
			return
		}
	}

	if traceID := firstNonEmptyHeader(headers, "X-Trace-Id", "x-trace-id"); traceID != "" {
		c.Header("X-Trace-Id", traceID)
	}

	normalizedBody, normalizedStatus := normalizeInvokeResponse(status, headers, raw)
	c.JSON(normalizedStatus, normalizedBody)
}

func invokeGatewayEndpoint(
	c *gin.Context,
	endpoint string,
	body []byte,
	scheme, token, apiKey, requestID, tenantUUID, mockModule string,
) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	if tenantUUID != "" {
		req.Header.Set("X-Tenant-UUID", tenantUUID)
	}
	if mockModule != "" {
		req.Header.Set("X-PX-Use-Mock", mockModule)
	}
	if scheme == "apikey" {
		if apiKey != "" {
			req.Header.Set("Authorization", "ApiKey "+apiKey)
			req.Header.Set("X-API-Key", apiKey)
		}
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-PowerX-Service-Token", token)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, raw, nil
}

func buildTenantInvocationPayload(rawBody []byte, tenantUUID string) ([]byte, error) {
	var req capabilityInvokeRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, err
	}
	capabilityID := strings.TrimSpace(req.CapabilityID)
	if capabilityID == "" {
		return nil, fmt.Errorf("capabilityId is required")
	}
	payload := map[string]any{
		"capability_id": capabilityID,
		"payload":       req.Payload,
	}
	if action := strings.TrimSpace(req.Action); action != "" {
		payload["action"] = action
	}
	if preferred := strings.TrimSpace(req.PreferredProtocol); preferred != "" {
		payload["preferred_protocol"] = preferred
	}
	if tenant := strings.TrimSpace(tenantUUID); tenant != "" {
		payload["tenant_uuid"] = tenant
	}
	if payload["payload"] == nil {
		payload["payload"] = map[string]any{}
	}
	return json.Marshal(payload)
}

func normalizeInvokeResponse(status int, headers http.Header, raw []byte) (gin.H, int) {
	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)

	traceID := pickString(
		pickFromMap(payload, "traceId"),
		pickFromMap(payload, "trace_id"),
		firstNonEmptyHeader(headers, "X-Trace-Id", "x-trace-id"),
	)

	if status >= http.StatusBadRequest {
		message := pickString(
			pickFromNestedMap(payload, "error", "message"),
			pickFromMap(payload, "message"),
			http.StatusText(status),
		)
		details := any(nil)
		if errObj, ok := payload["error"]; ok {
			details = errObj
		} else if len(payload) > 0 {
			details = payload
		}
		body := gin.H{
			"error": gin.H{
				"message": message,
				"details": details,
			},
		}
		if traceID != "" {
			body["traceId"] = traceID
		}
		return body, status
	}

	statusText := pickString(
		pickFromMap(payload, "status"),
		pickFromNestedMap(payload, "data", "status"),
		"success",
	)
	data := any(nil)
	switch {
	case payload == nil:
		data = map[string]any{}
	case payload["data"] != nil:
		data = payload["data"]
	case payload["result"] != nil:
		data = payload["result"]
	default:
		data = payload
	}
	result := gin.H{
		"status": statusText,
		"data":   data,
	}
	if traceID != "" {
		result["traceId"] = traceID
	}
	if len(payload) > 0 {
		result["raw"] = payload
	}
	return result, status
}

func pickFromMap(src map[string]any, key string) any {
	if src == nil {
		return nil
	}
	return src[key]
}

func pickFromNestedMap(src map[string]any, parent, key string) any {
	if src == nil {
		return nil
	}
	raw := src[parent]
	obj, ok := raw.(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	return obj[key]
}

func pickString(values ...any) string {
	for _, value := range values {
		if s, ok := value.(string); ok {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func firstNonEmptyHeader(headers http.Header, keys ...string) string {
	if headers == nil {
		return ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(headers.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func ensureAPIV1(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "/api/v1") {
		return base
	}
	if strings.HasSuffix(lower, "/api") {
		return base + "/v1"
	}
	if strings.Contains(lower, "/api/v1/") {
		return base
	}
	return base + "/api/v1"
}

func normalizeGatewayAuthScheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "apikey", "api_key", "api-key":
		return "apikey"
	case "bearer":
		return "bearer"
	default:
		return ""
	}
}

func resolveGatewayAuthScheme() string {
	explicit := normalizeGatewayAuthScheme(os.Getenv("PX_GATEWAY_AUTH_SCHEME"))
	if explicit != "" {
		return explicit
	}
	if strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY")) != "" {
		return "apikey"
	}
	return "bearer"
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val
		}
	}
	return ""
}

func respondPlaceholder(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":  "pending",
		"message": message,
	})
}
