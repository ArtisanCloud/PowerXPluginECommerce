package taskcenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pluginbootstrap "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/bootstrap"
)

var ErrTaskSubmitNotAvailable = fmt.Errorf("task submit not available")
var ErrTaskUpdateNotAvailable = fmt.Errorf("task update not available")

type TaskSubmitRequest struct {
	Type       string
	TenantUUID string
	Payload    map[string]any
	Metadata   map[string]any
}

type TaskSubmitResult struct {
	TaskID string
}

type TaskSubmitter interface {
	Submit(ctx context.Context, req TaskSubmitRequest) (*TaskSubmitResult, error)
}

type TaskUpdateRequest struct {
	TaskID      string
	TenantUUID  string
	Status      string
	Message     string
	DownloadURL string
	Metadata    map[string]any
	CompletedAt *time.Time
}

type TaskReporter interface {
	Update(ctx context.Context, req TaskUpdateRequest) error
}

type FrameworkTaskClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
	Path    string
}

func NewFrameworkTaskClient(baseURL, token string) *FrameworkTaskClient {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = strings.TrimSpace(resolveFrameworkBaseURL())
	}
	tok := strings.TrimSpace(token)
	if tok == "" {
		tok, _ = pluginbootstrap.ResolveToolToken()
	}
	return &FrameworkTaskClient{
		BaseURL: strings.TrimRight(base, "/"),
		Token:   tok,
		Client:  defaultHTTPClient(),
		Path:    "/api/v1/internal/tasks",
	}
}

func (c *FrameworkTaskClient) Submit(ctx context.Context, req TaskSubmitRequest) (*TaskSubmitResult, error) {
	if c == nil || strings.TrimSpace(c.BaseURL) == "" || strings.TrimSpace(c.Token) == "" {
		return nil, ErrTaskSubmitNotAvailable
	}
	taskType := strings.TrimSpace(req.Type)
	if taskType == "" {
		return nil, fmt.Errorf("task type is required")
	}

	tenantUUID := c.resolveTenant(req.TenantUUID)
	if tenantUUID == "" {
		return nil, fmt.Errorf("framework task tenant missing in token tid")
	}

	body := map[string]any{
		"type":        taskType,
		"tenant_uuid": tenantUUID,
	}
	if len(req.Metadata) > 0 {
		body["metadata"] = req.Metadata
	}
	if len(req.Payload) > 0 {
		body["payload"] = req.Payload
	}

	respBody, statusCode, err := c.doJSON(ctx, http.MethodPost, c.endpoint(""), tenantUUID, body)
	if err != nil {
		return nil, ErrTaskSubmitNotAvailable
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("framework task submit failed: status=%d body=%s", statusCode, strings.TrimSpace(string(respBody)))
	}

	taskID := extractTaskID(respBody)
	if strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("framework task submit missing taskId")
	}
	return &TaskSubmitResult{TaskID: taskID}, nil
}

func (c *FrameworkTaskClient) Update(ctx context.Context, req TaskUpdateRequest) error {
	if c == nil || strings.TrimSpace(c.BaseURL) == "" || strings.TrimSpace(c.Token) == "" {
		return ErrTaskUpdateNotAvailable
	}
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return fmt.Errorf("task status is required")
	}
	tenantUUID := c.resolveTenant(req.TenantUUID)
	if tenantUUID == "" {
		return fmt.Errorf("framework task tenant missing in token tid")
	}

	body := map[string]any{
		"status": status,
	}
	if msg := strings.TrimSpace(req.Message); msg != "" {
		body["message"] = msg
	}
	if downloadURL := strings.TrimSpace(req.DownloadURL); downloadURL != "" {
		body["downloadUrl"] = downloadURL
	}
	if len(req.Metadata) > 0 {
		body["metadata"] = req.Metadata
	}
	if req.CompletedAt != nil {
		body["completedAt"] = req.CompletedAt.UTC().Format(time.RFC3339)
	}

	attempts := []struct {
		method string
		url    string
	}{
		{method: http.MethodPatch, url: c.endpoint("/" + taskID)},
		{method: http.MethodPatch, url: c.endpoint("/" + taskID + "/progress")},
		{method: http.MethodPost, url: c.endpoint("/" + taskID + "/status")},
	}

	var lastErr error
	for _, attempt := range attempts {
		respBody, statusCode, err := c.doJSON(ctx, attempt.method, attempt.url, tenantUUID, body)
		if err != nil {
			lastErr = err
			continue
		}
		if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
			return nil
		}
		lastErr = fmt.Errorf("framework task update failed: status=%d body=%s", statusCode, strings.TrimSpace(string(respBody)))
		if statusCode != http.StatusNotFound && statusCode != http.StatusMethodNotAllowed {
			return lastErr
		}
	}

	if lastErr == nil {
		return ErrTaskUpdateNotAvailable
	}
	return lastErr
}

func (c *FrameworkTaskClient) resolveTenant(candidate string) string {
	tenantUUID := strings.TrimSpace(candidate)
	if tenantUUID == "" {
		if tid, ok := pluginbootstrap.ParseTenantIDFromJWT(c.Token); ok {
			tenantUUID = strings.TrimSpace(tid)
		}
	}
	return tenantUUID
}

func (c *FrameworkTaskClient) endpoint(suffix string) string {
	path := strings.TrimSpace(c.Path)
	if path == "" {
		path = "/api/v1/internal/tasks"
	}
	return strings.TrimRight(c.BaseURL, "/") + path + suffix
}

func (c *FrameworkTaskClient) doJSON(ctx context.Context, method, endpoint, tenantUUID string, payload map[string]any) ([]byte, int, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-PowerX-Tenant", tenantUUID)

	client := c.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()
	respBody, _ := io.ReadAll(response.Body)
	return respBody, response.StatusCode, nil
}

func extractTaskID(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	payload := map[string]any{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if taskID := readString(payload, "taskId", "task_id", "id"); taskID != "" {
		return taskID
	}
	if nested, ok := payload["data"].(map[string]any); ok {
		if taskID := readString(nested, "taskId", "task_id", "id"); taskID != "" {
			return taskID
		}
	}
	return ""
}

func readString(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			switch typed := value.(type) {
			case string:
				if trimmed := strings.TrimSpace(typed); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}
