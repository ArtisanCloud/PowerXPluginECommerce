package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	pluginconfig "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	powerxclient "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/grpc/client"
	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/sirupsen/logrus"
)

type SchedulerMode string

const (
	SchedulerModeLocal SchedulerMode = "local"
	SchedulerModeCoreX SchedulerMode = "corex"
	SchedulerModeDual  SchedulerMode = "dual"
)

type WorkerSpec struct {
	Name     string
	Interval time.Duration
	RunLoop  func(context.Context)
	RunOnce  func(context.Context) error
}

type RemoteJobSpec struct {
	Name         string
	ScheduleType string
	ScheduleExpr string
	Payload      map[string]any
}

type RemoteScheduler interface {
	Upsert(ctx context.Context, spec RemoteJobSpec) error
}

type SchedulerBridge struct {
	mode          SchedulerMode
	fallbackLocal bool
	remote        RemoteScheduler
	logger        *logrus.Entry
}

func NewSchedulerBridge(mode SchedulerMode, fallbackLocal bool, remote RemoteScheduler, logger *logrus.Entry) *SchedulerBridge {
	if logger == nil {
		logger = pxlogger.WithField("component", "scheduler-bridge")
	}
	if mode == "" {
		mode = SchedulerModeLocal
	}
	return &SchedulerBridge{mode: mode, fallbackLocal: fallbackLocal, remote: remote, logger: logger}
}

func NewSchedulerBridgeFromEnv(logger *logrus.Entry) *SchedulerBridge {
	mode := resolveSchedulerModeFromEnv()
	fallbackLocal := resolveSchedulerFallbackFromEnv()
	bridge := NewSchedulerBridge(mode, fallbackLocal, nil, logger)

	if mode == SchedulerModeCoreX || mode == SchedulerModeDual {
		client, err := NewCoreXSchedulerClientFromEnv(10 * time.Second)
		if err != nil {
			bridge.logger.WithError(err).WithField("mode", mode).Warn("scheduler bridge corex client unavailable")
		} else {
			bridge.remote = client
		}
	}

	bridge.logger.WithFields(logrus.Fields{
		"mode":           bridge.mode,
		"fallback_local": bridge.fallbackLocal,
		"remote_ready":   bridge.remote != nil,
	}).Info("scheduler bridge resolved")
	return bridge
}

func (b *SchedulerBridge) StartWorkers(ctx context.Context, start func(func() error), specs ...WorkerSpec) {
	if b == nil || start == nil || len(specs) == 0 {
		return
	}

	localScheduler := NewScheduler(b.logger.WithField("sub_component", "local-scheduler"))
	hasLocalScheduledJob := false

	for _, spec := range specs {
		if strings.TrimSpace(spec.Name) == "" || (spec.RunLoop == nil && spec.RunOnce == nil) {
			continue
		}
		runLocal := b.mode == SchedulerModeLocal || b.mode == SchedulerModeDual
		if b.mode == SchedulerModeCoreX || b.mode == SchedulerModeDual {
			if err := b.upsertRemoteJob(ctx, spec); err != nil {
				b.logger.WithError(err).WithFields(logrus.Fields{
					"worker": spec.Name,
					"mode":   b.mode,
				}).Warn("scheduler bridge remote registration failed")
				if b.mode == SchedulerModeCoreX {
					runLocal = b.fallbackLocal
				}
			} else if b.mode == SchedulerModeCoreX {
				runLocal = false
			}
		}

		if !runLocal {
			b.logger.WithFields(logrus.Fields{
				"worker": spec.Name,
				"mode":   b.mode,
			}).Info("scheduler bridge skip local worker runner")
			continue
		}

		if spec.RunOnce != nil {
			interval := spec.Interval
			if interval <= 0 {
				interval = time.Minute
			}
			worker := spec
			localScheduler.Register(NewJobFunc(worker.Name, interval, worker.RunOnce))
			hasLocalScheduledJob = true
			b.logger.WithFields(logrus.Fields{
				"worker":   worker.Name,
				"mode":     b.mode,
				"interval": interval.String(),
			}).Info("scheduler bridge register local scheduled worker")
			continue
		}

		worker := spec
		start(func() error {
			b.logger.WithFields(logrus.Fields{
				"worker":   worker.Name,
				"mode":     b.mode,
				"interval": worker.Interval.String(),
			}).Info("scheduler bridge start local loop worker")
			worker.RunLoop(ctx)
			return nil
		})
	}

	if hasLocalScheduledJob {
		start(func() error {
			localScheduler.Start(ctx)
			<-ctx.Done()
			return nil
		})
	}
}

func (b *SchedulerBridge) upsertRemoteJob(ctx context.Context, spec WorkerSpec) error {
	if b.remote == nil {
		return fmt.Errorf("remote scheduler not configured")
	}
	interval := spec.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	scheduleExpr := strconv.FormatInt(int64(interval.Seconds()), 10) + "s"
	return b.remote.Upsert(ctx, RemoteJobSpec{
		Name:         spec.Name,
		ScheduleType: "interval",
		ScheduleExpr: scheduleExpr,
		Payload: map[string]any{
			"plugin_action": "scheduler.trigger",
			"params": map[string]any{
				"worker": spec.Name,
			},
		},
	})
}

func resolveSchedulerModeFromEnv() SchedulerMode {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("POWERX_SCHEDULER_MODE")))
	switch raw {
	case string(SchedulerModeCoreX):
		return SchedulerModeCoreX
	case string(SchedulerModeDual):
		return SchedulerModeDual
	default:
		return SchedulerModeLocal
	}
}

func resolveSchedulerFallbackFromEnv() bool {
	raw := strings.TrimSpace(os.Getenv("POWERX_SCHEDULER_FALLBACK_LOCAL"))
	if raw == "" {
		return true
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func resolveCoreXSchedulerAuthorization(ctx context.Context) (string, error) {
	if apiKey := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY")); apiKey != "" {
		return "ApiKey " + apiKey, nil
	}

	grpcAddress := strings.TrimSpace(os.Getenv("POWERX_GRPC_UPSTREAM_ADDRESS"))
	if grpcAddress == "" {
		grpcAddress = strings.TrimSpace(os.Getenv("PX_GATEWAY_GRPC_TARGET"))
	}
	stsClientID := strings.TrimSpace(os.Getenv("POWERX_STS_CLIENT_ID"))
	stsClientSecret := strings.TrimSpace(os.Getenv("POWERX_STS_CLIENT_SECRET"))
	if grpcAddress == "" || stsClientID == "" || stsClientSecret == "" {
		return "", fmt.Errorf("scheduler bridge requires PX_GATEWAY_API_KEY or POWERX_GRPC_UPSTREAM_ADDRESS with POWERX_STS_CLIENT_ID/POWERX_STS_CLIENT_SECRET")
	}

	client, err := powerxclient.NewPowerXServiceClient(ctx, &pluginconfig.GRPCUpstream{
		Address:         grpcAddress,
		TenantUUID:      strings.TrimSpace(os.Getenv("POWERX_GRPC_UPSTREAM_TENANT_UUID")),
		STSClientID:     stsClientID,
		STSClientSecret: stsClientSecret,
		STSAudience:     firstNonEmpty(strings.TrimSpace(os.Getenv("POWERX_STS_AUDIENCE")), "powerx:api"),
		STSScope:        firstNonEmpty(strings.TrimSpace(os.Getenv("POWERX_STS_SCOPE")), "access"),
		STSTTL:          resolveSchedulerSTSTTL(),
		ConnectMode:     "lazy",
	})
	if err != nil {
		return "", err
	}
	token, err := client.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("scheduler bridge STS token is empty")
	}
	return "Bearer " + token, nil
}

func resolveSchedulerSTSTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("POWERX_STS_TTL"))
	if raw == "" {
		return 300 * time.Second
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 300 * time.Second
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
