package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type CoreXSchedulerClient struct {
	baseURL       string
	authorization string
	ownerID       string
	httpClient    *http.Client
}

func NewCoreXSchedulerClientFromEnv(timeout time.Duration) (*CoreXSchedulerClient, error) {
	authorization, err := resolveCoreXSchedulerAuthorization(context.Background())
	if err != nil {
		return nil, err
	}
	baseURL := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	baseURL = normalizeSchedulerBaseURL(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("scheduler bridge base url missing")
	}
	ownerID := strings.TrimSpace(os.Getenv("POWERX_PLUGIN_ID"))
	if ownerID == "" {
		ownerID = "com.powerx.plugins.ecommerce"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &CoreXSchedulerClient{
		baseURL:       baseURL,
		authorization: authorization,
		ownerID:       ownerID,
		httpClient:    &http.Client{Timeout: timeout},
	}, nil
}

func (c *CoreXSchedulerClient) Upsert(ctx context.Context, spec RemoteJobSpec) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf("scheduler corex client unavailable")
	}
	name := strings.TrimSpace(spec.Name)
	if name == "" {
		return fmt.Errorf("scheduler job name is required")
	}
	scheduleType := strings.TrimSpace(spec.ScheduleType)
	if scheduleType == "" {
		scheduleType = "interval"
	}
	scheduleExpr := strings.TrimSpace(spec.ScheduleExpr)
	if scheduleExpr == "" {
		return fmt.Errorf("scheduler schedule expr is required")
	}

	body := map[string]any{
		"owner_type":    "plugin",
		"owner_id":      c.ownerID,
		"name":          name,
		"schedule_type": scheduleType,
		"schedule_expr": scheduleExpr,
		"payload":       spec.Payload,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := c.baseURL + "/api/v1/admin/scheduler/jobs"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("scheduler upsert failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(rb)))
	}
	return nil
}

func normalizeSchedulerBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(strings.ToLower(base), "/api/v1") {
		base = strings.TrimSuffix(base, "/api/v1")
	}
	return base
}
