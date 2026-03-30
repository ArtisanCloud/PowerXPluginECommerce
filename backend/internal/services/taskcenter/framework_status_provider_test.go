package taskcenter

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFrameworkStatusProviderGetSuccess(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("X-PowerX-Tenant"); got != "tenant-1" {
			t.Fatalf("unexpected tenant header: %s", got)
		}
		if !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks/task-1") {
			t.Fatalf("unexpected request url: %s", r.URL.String())
		}
		body := `{"success":true,"data":{"taskId":"task-1","type":"import","status":"running","message":"ok","createdAt":"2026-02-09T10:00:00Z"}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	provider := NewFrameworkStatusProvider("http://gateway.local", "test-token")
	provider.Client = client
	job, err := provider.Get(context.Background(), "task-1", "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil || job.TaskID != "task-1" || job.Status != "running" {
		t.Fatalf("unexpected job: %#v", job)
	}
}

func TestFrameworkStatusProviderGetNotAvailableOnMissingConfig(t *testing.T) {
	provider := NewFrameworkStatusProvider("", "")
	_, err := provider.Get(context.Background(), "task-1", "tenant-1")
	if err != ErrTaskStatusNotAvailable {
		t.Fatalf("expected ErrTaskStatusNotAvailable, got %v", err)
	}
}
