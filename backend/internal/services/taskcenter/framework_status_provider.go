package taskcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type FrameworkStatusProvider struct {
	BaseURL string
	Token   string
	Client  *http.Client
	Path    string
}

type frameworkJobEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		TaskID      string         `json:"taskId"`
		Type        string         `json:"type"`
		Status      string         `json:"status"`
		Message     string         `json:"message"`
		DownloadURL string         `json:"downloadUrl"`
		CreatedAt   string         `json:"createdAt"`
		CompletedAt *string        `json:"completedAt"`
		Metadata    map[string]any `json:"metadata"`
	} `json:"data"`
	Message string `json:"message"`
}

func NewFrameworkStatusProvider(baseURL, token string) *FrameworkStatusProvider {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = strings.TrimSpace(resolveFrameworkBaseURL())
	}
	tok := strings.TrimSpace(token)
	return &FrameworkStatusProvider{
		BaseURL: strings.TrimRight(base, "/"),
		Token:   tok,
		Client:  defaultHTTPClient(),
		Path:    "/api/v1/internal/tasks/",
	}
}

func (p *FrameworkStatusProvider) Get(ctx context.Context, taskID string, tenantUUID string) (*JobStatus, error) {
	if p == nil {
		return nil, ErrTaskStatusNotAvailable
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, ErrTaskStatusNotAvailable
	}
	if p.BaseURL == "" || p.Token == "" {
		return nil, ErrTaskStatusNotAvailable
	}

	client := p.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	path := strings.TrimSpace(p.Path)
	if path == "" {
		path = "/api/v1/internal/tasks/"
	}
	endpoint := strings.TrimRight(p.BaseURL, "/") + path + url.PathEscape(taskID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(tenantUUID) != "" {
		req.Header.Set("X-PowerX-Tenant", strings.TrimSpace(tenantUUID))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrTaskStatusNotAvailable
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrTaskNotFound
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("framework status query failed: %d", resp.StatusCode)
	}

	var env frameworkJobEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, ErrTaskStatusNotAvailable
	}
	if !env.Success && strings.TrimSpace(env.Data.TaskID) == "" {
		return nil, ErrTaskNotFound
	}

	job := &JobStatus{
		TaskID:      env.Data.TaskID,
		Type:        env.Data.Type,
		Status:      env.Data.Status,
		Message:     env.Data.Message,
		DownloadURL: env.Data.DownloadURL,
		Metadata:    env.Data.Metadata,
	}
	if job.TaskID == "" {
		job.TaskID = taskID
	}
	if t, err := time.Parse(time.RFC3339, env.Data.CreatedAt); err == nil {
		job.CreatedAt = t
	}
	if env.Data.CompletedAt != nil {
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(*env.Data.CompletedAt)); err == nil {
			job.CompletedAt = &t
		}
	}
	return job, nil
}

func resolveFrameworkBaseURL() string {
	if base := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}
	if base := strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return ""
}
