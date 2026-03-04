package taskcenter

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFrameworkTaskClientSubmitSuccess(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("unexpected method: %s", got)
		}
		if !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks") {
			t.Fatalf("unexpected request url: %s", r.URL.String())
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("X-PowerX-Tenant"); got != "tenant-1" {
			t.Fatalf("unexpected tenant header: %s", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"type":"product_spu_import"`) {
			t.Fatalf("unexpected request body: %s", string(body))
		}
		responseBody := `{"success":true,"data":{"taskId":"task-123"}}`
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		}, nil
	})}

	taskClient := NewFrameworkTaskClient("http://gateway.local", "token-1")
	taskClient.Client = client

	result, err := taskClient.Submit(context.Background(), TaskSubmitRequest{
		Type:       "product_spu_import",
		TenantUUID: "tenant-1",
		Metadata:   map[string]any{"templateId": "default"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.TaskID != "task-123" {
		t.Fatalf("unexpected submit result: %#v", result)
	}
}

func TestFrameworkTaskClientUpdateSuccess(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Method; got != http.MethodPatch {
			t.Fatalf("unexpected method: %s", got)
		}
		if !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks/task-123") {
			t.Fatalf("unexpected request url: %s", r.URL.String())
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"status":"success"`) {
			t.Fatalf("unexpected request body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	taskClient := NewFrameworkTaskClient("http://gateway.local", "token-1")
	taskClient.Client = client
	completed := time.Now().UTC()
	err := taskClient.Update(context.Background(), TaskUpdateRequest{
		TaskID:      "task-123",
		TenantUUID:  "tenant-1",
		Status:      "success",
		Message:     "done",
		DownloadURL: "tmp/report.json",
		CompletedAt: &completed,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFrameworkTaskClientUpdateFallbackToStatusEndpoint(t *testing.T) {
	attempt := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempt++
		switch attempt {
		case 1:
			if r.Method != http.MethodPatch || !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks/task-123") {
				t.Fatalf("unexpected first attempt: %s %s", r.Method, r.URL.String())
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(`not found`)), Header: make(http.Header)}, nil
		case 2:
			if r.Method != http.MethodPatch || !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks/task-123/progress") {
				t.Fatalf("unexpected second attempt: %s %s", r.Method, r.URL.String())
			}
			return &http.Response{StatusCode: http.StatusMethodNotAllowed, Body: io.NopCloser(strings.NewReader(`no`)), Header: make(http.Header)}, nil
		default:
			if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.String(), "/api/v1/internal/tasks/task-123/status") {
				t.Fatalf("unexpected third attempt: %s %s", r.Method, r.URL.String())
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
		}
	})}

	taskClient := NewFrameworkTaskClient("http://gateway.local", "token-1")
	taskClient.Client = client
	err := taskClient.Update(context.Background(), TaskUpdateRequest{
		TaskID:     "task-123",
		TenantUUID: "tenant-1",
		Status:     "failed",
		Message:    "boom",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempt != 3 {
		t.Fatalf("unexpected attempts: %d", attempt)
	}
}

func TestFrameworkTaskClientSubmitMissingConfig(t *testing.T) {
	taskClient := NewFrameworkTaskClient("", "")
	_, err := taskClient.Submit(context.Background(), TaskSubmitRequest{Type: "product_spu_import"})
	if err == nil {
		t.Fatalf("expected submit error")
	}
}
