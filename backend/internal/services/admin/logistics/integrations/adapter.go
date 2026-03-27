package integrations

import (
	"context"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
)

// WaybillCreateInput is a normalized payload passed into provider adapters.
type WaybillCreateInput struct {
	OrderID     string
	ServiceCode string
	WaybillNo   string
	Metadata    map[string]any
}

// WaybillCreateResult is returned by provider adapters.
type WaybillCreateResult struct {
	WaybillNo string
	Status    string
	Metadata  map[string]any
}

// TrackingFetchInput is normalized query payload for provider tracking pull.
type TrackingFetchInput struct {
	WaybillNo string
	Limit     int
}

// TrackingFetchEvent describes one provider-side tracking node.
type TrackingFetchEvent struct {
	EventID     string
	Status      string
	Description string
	OccurredAt  *time.Time
	Payload     map[string]any
}

// Adapter defines provider integration extension points.
type Adapter interface {
	Provider() string
	TestConnectivity(ctx context.Context, carrier *LogisticsModel.Carrier) error
	CreateWaybill(ctx context.Context, carrier *LogisticsModel.Carrier, input WaybillCreateInput) (*WaybillCreateResult, error)
	FetchTracking(ctx context.Context, carrier *LogisticsModel.Carrier, input TrackingFetchInput) ([]TrackingFetchEvent, error)
	NormalizeTrackingStatus(status string) string
}

// NewAdapterRegistry returns built-in adapters keyed by provider code.
func NewAdapterRegistry() map[string]Adapter {
	return map[string]Adapter{
		"self":    NewSelfAdapter("self"),
		"sf":      NewGatewayAdapter("sf"),
		"jd":      NewGatewayAdapter("jd"),
		"cainiao": NewGatewayAdapter("cainiao"),
		"dhl":     NewGatewayAdapter("dhl"),
		"other":   NewGatewayAdapter("other"),
	}
}
