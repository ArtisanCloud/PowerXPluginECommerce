package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/gin-gonic/gin"
)

type fakeProvider struct {
	job *taskcenter.JobStatus
	err error
}

func (p *fakeProvider) Get(_ context.Context, _ string, _ string) (*taskcenter.JobStatus, error) {
	if p == nil {
		return nil, taskcenter.ErrTaskStatusNotAvailable
	}
	return p.job, p.err
}

func TestGetJobStatus_ReturnsFrameworkResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	h := &Handler{provider: &fakeProvider{job: &taskcenter.JobStatus{TaskID: "task-1", Status: "success", CreatedAt: now}}}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/jobs/task-1", nil)
	req = req.WithContext(authx.ContextWithTenantUUID(req.Context(), "tenant-1"))
	c.Request = req
	c.Params = gin.Params{{Key: "taskId", Value: "task-1"}}

	h.GetJobStatus(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if ok, _ := body["success"].(bool); !ok {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestGetJobStatus_NotFoundWhenUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{provider: &fakeProvider{err: taskcenter.ErrTaskNotFound}}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/jobs/task-2", nil)
	req = req.WithContext(authx.ContextWithTenantUUID(req.Context(), "tenant-1"))
	c.Request = req
	c.Params = gin.Params{{Key: "taskId", Value: "task-2"}}

	h.GetJobStatus(c)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetJobStatus_BadGatewayWhenFrameworkUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{provider: &fakeProvider{err: taskcenter.ErrTaskStatusNotAvailable}}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/jobs/task-3", nil)
	req = req.WithContext(authx.ContextWithTenantUUID(req.Context(), "tenant-1"))
	c.Request = req
	c.Params = gin.Params{{Key: "taskId", Value: "task-3"}}

	h.GetJobStatus(c)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d, body=%s", rec.Code, rec.Body.String())
	}
}
