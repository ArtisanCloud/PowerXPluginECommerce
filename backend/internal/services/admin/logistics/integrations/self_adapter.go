package integrations

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
)

// SelfAdapter is a safe default adapter used for local and mocked providers.
type SelfAdapter struct {
	provider string
}

func NewSelfAdapter(provider string) *SelfAdapter {
	p := strings.TrimSpace(strings.ToLower(provider))
	if p == "" {
		p = "self"
	}
	return &SelfAdapter{provider: p}
}

func (a *SelfAdapter) Provider() string {
	if a == nil {
		return "self"
	}
	return a.provider
}

func (a *SelfAdapter) TestConnectivity(_ context.Context, carrier *LogisticsModel.Carrier) error {
	if carrier == nil {
		return errors.New("carrier missing")
	}
	cfg := ParseConfigMap(carrier.Config)
	if force, ok := cfg["force_unreachable"].(bool); ok && force {
		return errors.New("provider forced unreachable by config")
	}
	if strings.TrimSpace(carrier.Status) == "disabled" {
		return errors.New("carrier is disabled")
	}
	return nil
}

func (a *SelfAdapter) CreateWaybill(_ context.Context, _ *LogisticsModel.Carrier, input WaybillCreateInput) (*WaybillCreateResult, error) {
	waybillNo := strings.TrimSpace(input.WaybillNo)
	if waybillNo == "" {
		prefix := strings.ToUpper(strings.ReplaceAll(a.Provider(), "-", ""))
		if prefix == "" {
			prefix = "SELF"
		}
		waybillNo = fmt.Sprintf("%s%s", prefix, time.Now().UTC().Format("20060102150405"))
	}
	return &WaybillCreateResult{
		WaybillNo: waybillNo,
		Status:    "created",
		Metadata: map[string]any{
			"provider": a.Provider(),
		},
	}, nil
}

func (a *SelfAdapter) FetchTracking(_ context.Context, _ *LogisticsModel.Carrier, _ TrackingFetchInput) ([]TrackingFetchEvent, error) {
	return nil, nil
}

func (a *SelfAdapter) NormalizeTrackingStatus(status string) string {
	val := strings.TrimSpace(strings.ToLower(status))
	if val == "" {
		return "in_transit"
	}
	return strings.ReplaceAll(val, "-", "_")
}
