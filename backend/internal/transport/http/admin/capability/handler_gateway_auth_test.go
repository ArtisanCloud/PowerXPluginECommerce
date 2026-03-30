package capability

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchCoreXCatalog_APIKeyModeSetsAPIKeyAuthorizationHeader(t *testing.T) {
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "apikey")
	t.Setenv("PX_GATEWAY_API_KEY", "k_test_123")
	t.Setenv("PX_TOOL_TOKEN", "replace-me-with-tool-token")

	var gotAuthorization string
	var gotServiceToken string
	var gotAPIKey string
	restore := swapDefaultTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuthorization = r.Header.Get("Authorization")
		gotServiceToken = r.Header.Get("X-PowerX-Service-Token")
		gotAPIKey = r.Header.Get("X-API-Key")
		body := io.NopCloser(strings.NewReader(`{"success":true,"data":[{"id":"com.corex.sample"}]}`))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       body,
			Request:    r,
		}, nil
	}))
	defer restore()

	t.Setenv("PX_GATEWAY_BASE_URL", "http://gateway.example")

	_, err := fetchCoreXCatalog(context.Background(), "corex")
	if err != nil {
		t.Fatalf("fetchCoreXCatalog returned error: %v", err)
	}
	if gotAuthorization != "ApiKey k_test_123" {
		t.Fatalf("expected ApiKey Authorization in apikey mode, got %q", gotAuthorization)
	}
	if gotServiceToken != "" {
		t.Fatalf("expected empty X-PowerX-Service-Token in apikey mode, got %q", gotServiceToken)
	}
	if gotAPIKey != "k_test_123" {
		t.Fatalf("expected X-API-Key to be set, got %q", gotAPIKey)
	}
}

func TestFetchCoreXCatalog_BearerModeUsesToolToken(t *testing.T) {
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "bearer")
	t.Setenv("PX_GATEWAY_API_KEY", "")
	t.Setenv("PX_TOOL_TOKEN", "token-demo")

	var gotAuthorization string
	var gotServiceToken string
	var gotAPIKey string
	restore := swapDefaultTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuthorization = r.Header.Get("Authorization")
		gotServiceToken = r.Header.Get("X-PowerX-Service-Token")
		gotAPIKey = r.Header.Get("X-API-Key")
		body := io.NopCloser(strings.NewReader(`{"success":true,"data":[{"id":"com.corex.sample"}]}`))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       body,
			Request:    r,
		}, nil
	}))
	defer restore()

	t.Setenv("PX_GATEWAY_BASE_URL", "http://gateway.example")

	_, err := fetchCoreXCatalog(context.Background(), "corex")
	if err != nil {
		t.Fatalf("fetchCoreXCatalog returned error: %v", err)
	}
	if gotAuthorization != "Bearer token-demo" {
		t.Fatalf("unexpected Authorization header: %q", gotAuthorization)
	}
	if gotServiceToken != "token-demo" {
		t.Fatalf("unexpected X-PowerX-Service-Token header: %q", gotServiceToken)
	}
	if gotAPIKey != "" {
		t.Fatalf("expected empty X-API-Key in bearer mode, got %q", gotAPIKey)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func swapDefaultTransport(next http.RoundTripper) func() {
	prev := http.DefaultTransport
	http.DefaultTransport = next
	return func() {
		http.DefaultTransport = prev
	}
}
