package taskbus

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	frameworkevent "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/event"
	"github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/eventbridge"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/sirupsen/logrus"
)

type FrameworkClientConfig struct {
	Mode           string
	Enabled        bool
	FallbackLocal  bool
	LocalQueueSize int

	BaseURL        string
	Token          string
	TenantUUID     string
	UserAgent      string
	Timeout        time.Duration
	SourcePlugin   string
	PayloadVersion string
}

type FrameworkClient struct {
	emitter eventbridge.Emitter
	local   *LocalClient
	logger  *logrus.Entry
	cfg     FrameworkClientConfig
}

func NewFrameworkClient(cfg FrameworkClientConfig, logger *logrus.Entry) (*FrameworkClient, error) {
	if logger == nil {
		logger = logrus.New().WithField("component", "taskbus-framework")
	}

	normalized := normalizeFrameworkConfig(cfg)
	factory, err := eventbridge.NewFactory(eventbridge.Config{
		Enabled:         normalized.Enabled,
		Mode:            normalized.Mode,
		FallbackToLocal: normalized.FallbackLocal,
		LocalQueueSize:  normalized.LocalQueueSize,
	})
	if err != nil {
		return nil, err
	}

	emitter, err := factory.NewEmitter()
	if err != nil {
		return nil, err
	}

	return &FrameworkClient{
		emitter: emitter,
		local:   NewLocalClient(logger.WithField("sub_component", "framework-local-subscriber")),
		logger:  logger,
		cfg:     normalized,
	}, nil
}

func (c *FrameworkClient) Publish(ctx context.Context, evt Event) error {
	if c == nil || c.emitter == nil {
		return nil
	}
	ev, err := c.toFrameworkEvent(ctx, evt)
	if err != nil {
		return err
	}
	if err := c.emitter.Emit(ctx, ev); err != nil {
		return err
	}
	if c.local != nil {
		_ = c.local.Publish(ctx, evt)
	}
	return nil
}

func (c *FrameworkClient) Subscribe(topic string, handler Handler) error {
	if c == nil || c.local == nil {
		return nil
	}
	return c.local.Subscribe(topic, handler)
}

func (c *FrameworkClient) Close(context.Context) error { return nil }

func (c *FrameworkClient) toFrameworkEvent(ctx context.Context, evt Event) (frameworkevent.Event, error) {
	topic := strings.TrimSpace(evt.Topic)
	if topic == "" {
		return frameworkevent.Event{}, nil
	}

	tenantUUID := strings.TrimSpace(evt.Metadata["tenant_uuid"])
	if tenantUUID == "" {
		if tid, ok := authx.TenantUUIDFromContext(ctx); ok {
			tenantUUID = strings.TrimSpace(tid)
		}
	}
	if tenantUUID == "" {
		tenantUUID = strings.TrimSpace(c.cfg.TenantUUID)
	}
	requestID := strings.TrimSpace(evt.Metadata["request_id"])
	traceID := strings.TrimSpace(evt.Metadata["trace_id"])
	sourcePlugin := strings.TrimSpace(evt.Metadata["source_plugin"])
	if sourcePlugin == "" {
		sourcePlugin = c.cfg.SourcePlugin
	}
	payloadVersion := strings.TrimSpace(evt.Metadata["payload_version"])
	if payloadVersion == "" {
		payloadVersion = c.cfg.PayloadVersion
	}

	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return frameworkevent.Event{}, err
	}

	return frameworkevent.Event{
		Topic: frameworkevent.Topic(topic),
		Meta: frameworkevent.Meta{
			TenantUUID:     tenantUUID,
			RequestID:      requestID,
			TraceID:        traceID,
			SourcePlugin:   sourcePlugin,
			PayloadVersion: payloadVersion,
			OccurredAt:     time.Now().UTC(),
		},
		Payload: payload,
	}, nil
}

func normalizeFrameworkConfig(cfg FrameworkClientConfig) FrameworkClientConfig {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		if strings.TrimSpace(os.Getenv("POWERX_PROXY")) == "1" {
			mode = "taskbus"
		} else {
			mode = "local"
		}
	}

	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		if v := strings.TrimSpace(os.Getenv("PX_TOOL_TOKEN")); v != "" {
			token = v
		} else if v := strings.TrimSpace(os.Getenv("PX_PLUGIN_TOOL_TOKEN")); v != "" {
			token = v
		} else {
			token = strings.TrimSpace(os.Getenv("POWERX_AUTH_TOKEN"))
		}
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	}
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	baseURL = normalizeHostBaseURL(baseURL)

	if cfg.LocalQueueSize <= 0 {
		cfg.LocalQueueSize = 1024
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if strings.TrimSpace(cfg.SourcePlugin) == "" {
		cfg.SourcePlugin = "com.powerx.plugins.ecommerce"
	}
	if strings.TrimSpace(cfg.PayloadVersion) == "" {
		cfg.PayloadVersion = "v1"
	}
	if !cfg.Enabled {
		cfg.Enabled = true
	}
	if !cfg.FallbackLocal {
		cfg.FallbackLocal = true
	}

	cfg.Mode = mode
	cfg.Token = token
	cfg.BaseURL = baseURL
	return cfg
}

func normalizeHostBaseURL(raw string) string {
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
