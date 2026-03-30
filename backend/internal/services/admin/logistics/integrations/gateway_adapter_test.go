package integrations

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestApplyAuthHeaders_DefaultAPIKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	cfg := map[string]any{"api_key": "k-123"}
	applyAuthHeaders(req, cfg)
	require.Equal(t, "k-123", req.Header.Get("X-API-Key"))
	require.Equal(t, "", req.Header.Get("Authorization"))
}

func TestApplyAuthHeaders_BearerAndBasic(t *testing.T) {
	reqBearer := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	applyAuthHeaders(reqBearer, map[string]any{"auth_scheme": "bearer", "tool_token": "t-abc"})
	require.Equal(t, "Bearer t-abc", reqBearer.Header.Get("Authorization"))

	reqBasic := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	applyAuthHeaders(reqBasic, map[string]any{"auth_scheme": "basic", "basic_username": "u", "basic_password": "p"})
	require.Equal(t, "Basic dTpw", reqBasic.Header.Get("Authorization"))
}

func TestGatewayAdapter_CreateWaybillRetryAndParse(t *testing.T) {
	attempts := 0
	adapter := NewGatewayAdapter("sf")
	adapter.client = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		require.Equal(t, "Bearer tok-1", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		if attempts == 1 {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader(`{"error":"temporary"}`)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"data":{"waybill_no":"SF888","status":"created"}}`)),
			Header:     make(http.Header),
		}, nil
	})}

	carrier := &LogisticsModel.Carrier{
		Config: datatypes.JSON([]byte(`{
			"gateway_base_url":"http://gateway.local",
			"gateway_retry_count":1,
			"auth_scheme":"bearer",
			"tool_token":"tok-1"
		}`)),
	}

	res, err := adapter.CreateWaybill(context.Background(), carrier, WaybillCreateInput{
		OrderID:     "o1",
		ServiceCode: "std",
	})
	require.NoError(t, err)
	require.Equal(t, 2, attempts)
	require.Equal(t, "SF888", res.WaybillNo)
	require.Equal(t, "created", res.Status)
	require.Equal(t, "gateway", res.Metadata["source"])
}

func TestGatewayAdapter_FallbackWhenGatewayMissing(t *testing.T) {
	adapter := NewGatewayAdapter("jd")
	carrier := &LogisticsModel.Carrier{Status: "active"}
	res, err := adapter.CreateWaybill(context.Background(), carrier, WaybillCreateInput{})
	require.NoError(t, err)
	require.Contains(t, res.WaybillNo, "JD")
}

func TestGatewayAdapter_FetchTracking_ParseDataItemsAndLimitQuery(t *testing.T) {
	adapter := NewGatewayAdapter("sf")
	adapter.client = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		require.Equal(t, "/api/v1/integration/logistics/providers/sf/waybills/WB-100/tracking", r.URL.Path)
		require.Equal(t, "2", r.URL.Query().Get("limit"))
		require.Equal(t, "k-1", r.Header.Get("X-API-Key"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"data":{"items":[{"event_id":"E1","status":"in_transit","description":"scan","occurred_at":"2026-03-01T10:00:00Z","payload":{"city":"SH"}},{"event_id":"E2","status":"delivered","description":"signed","occurred_at":"2026-03-02T12:30:00Z"}]}}`,
			)),
			Header: make(http.Header),
		}, nil
	})}
	carrier := &LogisticsModel.Carrier{
		Config: datatypes.JSON([]byte(`{"gateway_base_url":"http://gateway.local","api_key":"k-1"}`)),
	}

	events, err := adapter.FetchTracking(context.Background(), carrier, TrackingFetchInput{
		WaybillNo: "WB-100",
		Limit:     2,
	})
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, "E1", events[0].EventID)
	require.Equal(t, "in_transit", events[0].Status)
	require.Equal(t, "SH", events[0].Payload["city"])
	require.NotNil(t, events[0].OccurredAt)
	require.Equal(t, "E2", events[1].EventID)
	require.Equal(t, "delivered", events[1].Status)
}

func TestGatewayAdapter_FetchTracking_RetryOn5xx(t *testing.T) {
	attempts := 0
	adapter := NewGatewayAdapter("jd")
	adapter.client = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader(`{"error":"temporary"}`)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"items":[{"event_id":"E3","status":"signed","description":"ok","occurred_at":"2026-03-03T08:00:00Z"}]}`,
			)),
			Header: make(http.Header),
		}, nil
	})}
	carrier := &LogisticsModel.Carrier{
		Config: datatypes.JSON([]byte(`{"gateway_base_url":"http://gateway.local","gateway_retry_count":1}`)),
	}

	events, err := adapter.FetchTracking(context.Background(), carrier, TrackingFetchInput{WaybillNo: "JD-1"})
	require.NoError(t, err)
	require.Equal(t, 2, attempts)
	require.Len(t, events, 1)
	require.Equal(t, "E3", events[0].EventID)
}

func TestGatewayAdapter_FetchTracking_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	adapter := NewGatewayAdapter("dhl")
	adapter.client = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{"error":"unauthorized"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	carrier := &LogisticsModel.Carrier{
		Config: datatypes.JSON([]byte(`{"gateway_base_url":"http://gateway.local","gateway_retry_count":3}`)),
	}

	events, err := adapter.FetchTracking(context.Background(), carrier, TrackingFetchInput{WaybillNo: "DHL-1"})
	require.Error(t, err)
	require.Nil(t, events)
	require.Equal(t, 1, attempts)
	require.Contains(t, err.Error(), "status=401")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
