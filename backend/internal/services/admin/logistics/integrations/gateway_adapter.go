package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
)

// GatewayAdapter forwards provider actions to an external logistics gateway.
type GatewayAdapter struct {
	provider string
	client   *http.Client
	fallback *SelfAdapter
}

func NewGatewayAdapter(provider string) *GatewayAdapter {
	p := strings.TrimSpace(strings.ToLower(provider))
	if p == "" {
		p = "other"
	}
	return &GatewayAdapter{
		provider: p,
		client:   &http.Client{Timeout: 3 * time.Second},
		fallback: NewSelfAdapter(p),
	}
}

func (a *GatewayAdapter) Provider() string {
	if a == nil {
		return "other"
	}
	return a.provider
}

func (a *GatewayAdapter) TestConnectivity(ctx context.Context, carrier *LogisticsModel.Carrier) error {
	if carrier == nil {
		return errors.New("carrier missing")
	}
	cfg := ParseConfigMap(carrier.Config)
	baseURL := strings.TrimSpace(anyString(cfg["gateway_base_url"]))
	if baseURL == "" {
		return a.fallback.TestConnectivity(ctx, carrier)
	}
	path := strings.TrimSpace(anyString(cfg["gateway_connectivity_path"]))
	if path == "" {
		path = "/api/v1/integration/logistics/providers/{provider}/connectivity"
	}
	url := joinGatewayURL(baseURL, strings.ReplaceAll(path, "{provider}", a.Provider()))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	applyAuthHeaders(req, cfg)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("gateway connectivity status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func (a *GatewayAdapter) CreateWaybill(ctx context.Context, carrier *LogisticsModel.Carrier, input WaybillCreateInput) (*WaybillCreateResult, error) {
	cfg := map[string]any{}
	if carrier != nil {
		cfg = ParseConfigMap(carrier.Config)
	}
	baseURL := strings.TrimSpace(anyString(cfg["gateway_base_url"]))
	if baseURL == "" {
		return a.fallback.CreateWaybill(ctx, carrier, input)
	}
	path := strings.TrimSpace(anyString(cfg["gateway_create_waybill_path"]))
	if path == "" {
		path = "/api/v1/integration/logistics/providers/{provider}/waybills"
	}
	url := joinGatewayURL(baseURL, strings.ReplaceAll(path, "{provider}", a.Provider()))

	payload := map[string]any{
		"provider":     a.Provider(),
		"order_id":     strings.TrimSpace(input.OrderID),
		"service_code": strings.TrimSpace(input.ServiceCode),
		"waybill_no":   strings.TrimSpace(input.WaybillNo),
		"metadata":     input.Metadata,
	}
	bodyBytes, _ := json.Marshal(payload)
	retry := int(anyFloat(cfg["gateway_retry_count"]))
	if retry < 0 {
		retry = 0
	}
	backoffMS := int(anyFloat(cfg["gateway_retry_backoff_ms"]))
	if backoffMS <= 0 {
		backoffMS = 120
	}

	var lastErr error
	for attempt := 0; attempt <= retry; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		applyAuthHeaders(req, cfg)

		resp, err := a.client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			data, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
			resp.Body.Close()
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				return parseGatewayWaybillResult(input, data, a.Provider()), nil
			}
			lastErr = fmt.Errorf("gateway create waybill status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(data)))
			if resp.StatusCode < 500 || resp.StatusCode >= 600 {
				break
			}
		}
		if attempt < retry {
			time.Sleep(time.Duration(backoffMS) * time.Millisecond)
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("gateway create waybill failed")
}

func (a *GatewayAdapter) FetchTracking(ctx context.Context, carrier *LogisticsModel.Carrier, input TrackingFetchInput) ([]TrackingFetchEvent, error) {
	cfg := map[string]any{}
	if carrier != nil {
		cfg = ParseConfigMap(carrier.Config)
	}
	baseURL := strings.TrimSpace(anyString(cfg["gateway_base_url"]))
	if baseURL == "" {
		return a.fallback.FetchTracking(ctx, carrier, input)
	}
	waybillNo := strings.TrimSpace(input.WaybillNo)
	if waybillNo == "" {
		return nil, errors.New("waybill_no required")
	}
	path := strings.TrimSpace(anyString(cfg["gateway_tracking_path"]))
	if path == "" {
		path = "/api/v1/integration/logistics/providers/{provider}/waybills/{waybill_no}/tracking"
	}
	path = strings.ReplaceAll(path, "{provider}", a.Provider())
	path = strings.ReplaceAll(path, "{waybill_no}", waybillNo)
	url := joinGatewayURL(baseURL, path)
	if input.Limit > 0 {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		url = fmt.Sprintf("%s%slimit=%d", url, sep, input.Limit)
	}

	retry := int(anyFloat(cfg["gateway_retry_count"]))
	if retry < 0 {
		retry = 0
	}
	backoffMS := int(anyFloat(cfg["gateway_retry_backoff_ms"]))
	if backoffMS <= 0 {
		backoffMS = 120
	}

	var lastErr error
	for attempt := 0; attempt <= retry; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		applyAuthHeaders(req, cfg)
		resp, err := a.client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			data, _ := io.ReadAll(io.LimitReader(resp.Body, 24*1024))
			resp.Body.Close()
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				return parseGatewayTrackingEvents(data), nil
			}
			lastErr = fmt.Errorf("gateway tracking status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(data)))
			if resp.StatusCode < 500 || resp.StatusCode >= 600 {
				break
			}
		}
		if attempt < retry {
			time.Sleep(time.Duration(backoffMS) * time.Millisecond)
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("gateway tracking fetch failed")
}

func (a *GatewayAdapter) NormalizeTrackingStatus(status string) string {
	return a.fallback.NormalizeTrackingStatus(status)
}

func parseGatewayWaybillResult(input WaybillCreateInput, raw []byte, provider string) *WaybillCreateResult {
	result := &WaybillCreateResult{
		WaybillNo: strings.TrimSpace(input.WaybillNo),
		Status:    "created",
		Metadata: map[string]any{
			"provider": provider,
			"source":   "gateway",
		},
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return result
	}
	if data, ok := payload["data"].(map[string]any); ok {
		payload = data
	}
	if v := strings.TrimSpace(anyString(payload["waybill_no"])); v != "" {
		result.WaybillNo = v
	}
	if v := strings.TrimSpace(anyString(payload["waybillNo"])); v != "" {
		result.WaybillNo = v
	}
	if v := strings.TrimSpace(anyString(payload["status"])); v != "" {
		result.Status = strings.ToLower(v)
	}
	result.Metadata["raw"] = payload
	return result
}

func parseGatewayTrackingEvents(raw []byte) []TrackingFetchEvent {
	if len(raw) == 0 {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	if data, ok := payload["data"].(map[string]any); ok {
		if items, ok := data["items"].([]any); ok {
			return normalizeGatewayTrackingItems(items)
		}
	}
	if items, ok := payload["items"].([]any); ok {
		return normalizeGatewayTrackingItems(items)
	}
	return nil
}

func normalizeGatewayTrackingItems(items []any) []TrackingFetchEvent {
	if len(items) == 0 {
		return nil
	}
	out := make([]TrackingFetchEvent, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		eventID := strings.TrimSpace(anyString(row["event_id"]))
		if eventID == "" {
			eventID = strings.TrimSpace(anyString(row["eventId"]))
		}
		status := strings.TrimSpace(anyString(row["status"]))
		description := strings.TrimSpace(anyString(row["description"]))
		var occurredAt *time.Time
		occurredRaw := strings.TrimSpace(anyString(row["occurred_at"]))
		if occurredRaw == "" {
			occurredRaw = strings.TrimSpace(anyString(row["occurredAt"]))
		}
		if occurredRaw != "" {
			if t, err := time.Parse(time.RFC3339, occurredRaw); err == nil {
				occurredAt = &t
			}
		}
		payload := map[string]any{}
		if v, ok := row["payload"].(map[string]any); ok {
			payload = v
		}
		out = append(out, TrackingFetchEvent{
			EventID:     eventID,
			Status:      status,
			Description: description,
			OccurredAt:  occurredAt,
			Payload:     payload,
		})
	}
	return out
}

func applyAuthHeaders(req *http.Request, cfg map[string]any) {
	authScheme := strings.TrimSpace(strings.ToLower(anyString(cfg["auth_scheme"])))
	if authScheme == "" {
		authScheme = strings.TrimSpace(strings.ToLower(anyString(cfg["gateway_auth_scheme"])))
	}
	if authScheme == "" {
		authScheme = "api_key"
	}

	apiKey := strings.TrimSpace(anyString(cfg["api_key"]))
	if apiKey == "" {
		apiKey = strings.TrimSpace(anyString(cfg["gateway_api_key"]))
	}
	token := strings.TrimSpace(anyString(cfg["token"]))
	if token == "" {
		token = strings.TrimSpace(anyString(cfg["tool_token"]))
	}
	if token == "" {
		token = strings.TrimSpace(anyString(cfg["gateway_token"]))
	}

	switch authScheme {
	case "bearer", "token":
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username := strings.TrimSpace(anyString(cfg["basic_username"]))
		password := strings.TrimSpace(anyString(cfg["basic_password"]))
		if username != "" || password != "" {
			cred := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			req.Header.Set("Authorization", "Basic "+cred)
		}
	default:
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		if req.Header.Get("Authorization") == "" && token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
}

func joinGatewayURL(baseURL, path string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	suffix := "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	return base + suffix
}

func anyFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}
