package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	fwbootstrap "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/bootstrap"
	"github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/manifest"
	fwrouter "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/router"
	pluginbootstrap "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/bootstrap"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	dbpkg "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/db"
	marketplacerepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/marketplace"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/plugin"
	grpcserver "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/grpc/server"
	channelmasterjobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/jobs/channel/master"
	integrationjobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/jobs/integration"
	marketplacejobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/jobs/marketplace"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	manifestx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/manifestx"
	adminmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/admin_console"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/auth"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	opsmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/operations"
	pluginrouter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/router"
	eventfabricruntime "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/eventfabric"
	httpserver "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/server"
	agent "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/authproxy"
	iamservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
	marketplacesvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/marketplace"
	recommendation "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/recommendation"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/utils"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/taskbus"
	wstransport "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket"
	wstransportbus "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket/bus"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	if os.Getenv("CONFIG_PATH") == "" && os.Getenv("POWERX_PLUGIN_CONFIG_DIR") != "" {
		os.Setenv("CONFIG_PATH", os.Getenv("POWERX_PLUGIN_CONFIG_DIR"))
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志隐私掩码规则
	masking := cfg.SecurityBaselineConfig().MaskingRules
	if len(masking.PIIFields) > 0 {
		placeholder := masking.LogRedaction.Placeholder
		logger.ConfigurePrivacyMasker(masking.PIIFields, placeholder)
	}

	manifestPath, executionPath := eventfabricruntime.ResolvePaths()
	if manifestPath == "" || executionPath == "" {
		logger.WithFields(logger.Fields{
			"plugin_yaml":       manifestPath,
			"event_fabric_yaml": executionPath,
			"strict_validation": strings.TrimSpace(os.Getenv("POWERX_EVENT_FABRIC_STRICT")) != "",
		}).Warn("event fabric alignment skipped: declaration files not found")
	} else {
		manifestTopics, mErr := eventfabricruntime.LoadManifestTopics(manifestPath)
		executionTopics, eErr := eventfabricruntime.LoadExecutionTopics(executionPath)
		if mErr != nil || eErr != nil {
			logger.WithFields(logger.Fields{
				"manifest_err":  errorString(mErr),
				"execution_err": errorString(eErr),
				"plugin_yaml":   manifestPath,
				"event_fabric":  executionPath,
			}).Warn("event fabric alignment check failed to parse")
		} else {
			localTopics := make([]string, 0, len(executionTopics))
			for _, topic := range executionTopics {
				localTopics = append(localTopics, topic.Topic)
			}
			wstransportbus.DefaultTopicRegistry.Register(localTopics)
			if err := eventfabricruntime.ValidateConsistency(manifestTopics, executionTopics); err != nil {
				if strings.TrimSpace(os.Getenv("POWERX_EVENT_FABRIC_STRICT")) == "1" {
					logger.WithError(err).Fatal("event fabric declaration mismatch")
				}
				logger.WithError(err).Warn("event fabric declaration mismatch")
			} else {
				logger.WithField("topics", len(localTopics)).Info("event fabric declaration aligned")
			}
		}
	}

	// ★ 在这里把 HTTP/GRPC 的占位符先解析掉（一定要在起服务之前）
	//   - HTTP 用 PORT（由 PowerX 的 supervisor 注入）
	cfg.Server.BindAddr = utils.ResolveDynamicAddr(cfg.Server.BindAddr, "PORT")

	//   - gRPC 用 POWERX_GRPC_PORT（由 PowerX 的 Enable 阶段注入）
	if cfg.GRPCServer != nil {
		// 如果你的字段叫 Addr，就把下一行改成：cfg.GRPCServer.Addr = resolveDynamicAddr(cfg.GRPCServer.Addr, "POWERX_GRPC_PORT")
		cfg.GRPCServer.Addr = utils.ResolveDynamicAddr(cfg.GRPCServer.Addr, "POWERX_GRPC_PORT")
	}

	// 初始化插件
	queryDB, err := pluginbootstrap.BootstrapPlugin(ctx, cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to bootstrap plugin")
	}
	if cfg != nil && cfg.Server.DevMode {
		if n, err := pluginbootstrap.RepairDraftSKUsForPublishedSPUs(ctx, queryDB); err != nil {
			logger.WithError(err).Warn("Dev repair: failed to reconcile SKU statuses for published SPUs")
		} else if n > 0 {
			logger.WithField("rows", n).Info("Dev repair: reconciled SKU statuses for published SPUs")
		}
	}

	// 在初始化 gRPC 客户端之前，尝试从本地数据库加载租户凭证（若存在），以便通过 STS 获取短期令牌
	if cfg.GRPCUpstream != nil && strings.TrimSpace(cfg.GRPCUpstream.TenantUUID) != "" {
		// 延迟依赖：仅当配置未提供 STS client 时，尝试 DB 加载；若配置已有，则优先生效
		if cfg.GRPCUpstream.STSClientID == "" || cfg.GRPCUpstream.STSClientSecret == "" {
			repo := repository.NewCredentialsRepository(queryDB)
			svc := agent.NewCredentialService(cfg, repo)
			if cid, sec, err := svc.LoadDecryptedCredentials(rootCtx, cfg.GRPCUpstream.TenantUUID, app.PluginID); err == nil {
				cfg.GRPCUpstream.STSClientID = cid
				cfg.GRPCUpstream.STSClientSecret = sec
				logger.Info("Loaded STS credentials for tenant from DB")
			} else {
				logger.WithError(err).Warn("No DB-stored credentials found or failed to decrypt; will rely on config/env if provided")
			}
		}
	}

	iamResolver := pluginbootstrap.NewIAMResolver(cfg)
	runtimeDecision := pluginbootstrap.ResolveRuntimeModeDecision(cfg, iamResolver.Mode().String(), iamResolver.Source())
	wsDriver := cfg.ResolveWebSocketDriver()
	eventTopicDriver := cfg.ResolveEventTopicDriver()
	taskDriver := cfg.ResolveTaskDriver()
	cacheDriver := cfg.ResolveCacheDriver()
	upstreamAddr := ""
	upstreamTenant := ""
	if cfg.GRPCUpstream != nil {
		upstreamAddr = strings.TrimSpace(cfg.GRPCUpstream.Address)
		upstreamTenant = strings.TrimSpace(cfg.GRPCUpstream.TenantUUID)
	}
	logger.WithFields(logger.Fields{
		"matrix":                     "IAMMode × POWERX_PROXY",
		"iam_input":                  runtimeDecision.IAMInput,
		"iam_mode":                   runtimeDecision.IAMMode,
		"iam_source":                 runtimeDecision.IAMSource,
		"powerx_proxy":               runtimeDecision.PowerXProxy,
		"capability_route":           runtimeDecision.CapabilityRoute,
		"ws_route":                   runtimeDecision.WSRoute,
		"outbound_token_source":      runtimeDecision.OutboundTokenSource,
		"gateway_readiness":          runtimeDecision.GatewayReady,
		"gateway_token_tenant_tid":   runtimeDecision.TokenTenantID,
		"gateway_upstream_address":   upstreamAddr,
		"gateway_upstream_tenant_id": upstreamTenant,
		"ws_driver":                  wsDriver,
		"event_topic_driver":         eventTopicDriver,
		"task_driver":                taskDriver,
		"cache_driver":               cacheDriver,
	}).Info("runtime mode decision resolved")
	auth.ObserveMode(iamResolver.Mode().String())

	var authClient *authproxy.DelegatedClient
	var localIAM iamservice.IAMDirectory
	if iamResolver.Mode() == iamservice.IAMModeDelegated {
		client, err := authproxy.NewDelegatedClient("", "")
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize delegated auth proxy; auth endpoints will be unavailable")
		} else {
			authClient = client
		}
	} else {
		dir, err := iamservice.NewLocalDirectory(queryDB, cfg)
		if err != nil {
			logger.WithError(err).Fatal("Failed to initialize local IAM directory")
		}
		localIAM = dir
	}

	// 初始化 PowerX gRPC Client 客户端
	pxc := pluginbootstrap.BootstrapGRPCClient(rootCtx, cfg.GRPCUpstream)

	taxLogger := logger.WithField("component", "tax_provider_client")
	taxClient, err := marketplacesvc.NewTaxProviderClient(cfg, nil, taxLogger)
	if err != nil {
		taxLogger.WithError(err).Warn("Tax provider client initialization failed")
	}

	var licenseCache marketplacesvc.LicenseCache
	cacheCfg := cfg.LicenseCacheConfig()
	cacheLogger := logger.WithField("component", "marketplace_license_cache")
	cacheProvider := strings.ToLower(strings.TrimSpace(cacheCfg.Provider))
	switch cacheDriver {
	case config.RuntimeDriverFramework:
		cacheLogger.WithField("driver", cacheDriver).Info("license cache uses framework/host side")
	default:
		switch cacheProvider {
		case "redis":
			if lc, err := marketplacesvc.NewRedisLicenseCache(cacheCfg.RedisURL, cacheCfg.KeyPrefix, cacheLogger); err != nil {
				cacheLogger.WithError(err).Warn("license cache initialization failed")
			} else {
				licenseCache = lc
			}
		case "", "memory", "local":
			licenseCache = marketplacesvc.NewMemoryLicenseCache(cacheLogger)
		default:
			cacheLogger.WithField("provider", cacheProvider).Warn("unsupported license cache provider, fallback to memory")
			licenseCache = marketplacesvc.NewMemoryLicenseCache(cacheLogger)
		}
	}

	var taskBusClient taskbus.Client
	taskBusEnabled := cfg.TaskBusEnabled()
	taskBusAdapter := cfg.TaskBusAdapter()
	if taskBusAdapter == "" {
		taskBusAdapter = taskDriver
	}
	if taskBusAdapter == config.RuntimeDriverFramework {
		taskBusAdapter = "framework"
	} else if taskBusAdapter == config.RuntimeDriverLocal {
		taskBusAdapter = "local"
	}
	if !taskBusEnabled && !runtimeDecision.PowerXProxy {
		// Standalone 默认启用本地 TaskBus，避免未配置时退化为 noop 导致前端 WS 订阅后收不到事件。
		taskBusEnabled = true
		if strings.TrimSpace(taskBusAdapter) == "" {
			taskBusAdapter = "local"
		}
		logger.WithFields(logger.Fields{
			"powerx_proxy": runtimeDecision.PowerXProxy,
			"adapter":      taskBusAdapter,
		}).Info("taskbus auto-enabled for standalone mode")
	}
	if !taskBusEnabled && taskBusAdapter == "framework" {
		taskBusEnabled = true
	}

	if taskBusEnabled {
		switch taskBusAdapter {
		case "framework", "powerx", "host":
			frameworkTaskBus, err := taskbus.NewFrameworkClient(taskbus.FrameworkClientConfig{
				Mode:           "taskbus",
				Enabled:        true,
				FallbackLocal:  true,
				LocalQueueSize: 1024,
				SourcePlugin:   app.PluginID,
				PayloadVersion: "v1",
			}, logger.WithField("component", "taskbus-framework"))
			if err != nil {
				logger.WithError(err).Warn("failed to initialize framework taskbus adapter, falling back to local")
				taskBusClient = taskbus.NewLocalClient(logger.WithField("component", "taskbus-local"))
			} else {
				taskBusClient = frameworkTaskBus
			}
		case "", "local":
			taskBusClient = taskbus.NewLocalClient(logger.WithField("component", "taskbus-local"))
		default:
			logger.WithField("adapter", taskBusAdapter).Warn("unsupported taskbus adapter, falling back to noop")
			taskBusClient = taskbus.NewNoopClient()
		}
	} else {
		taskBusClient = taskbus.NewNoopClient()
	}

	logger.WithFields(logger.Fields{
		"enabled": taskBusEnabled,
		"adapter": taskBusAdapter,
	}).Info("taskbus resolved")

	deps := &app.Deps{
		DB:                  queryDB,
		Ctx:                 rootCtx,
		PowerXClient:        pxc,
		Config:              cfg,
		TaxProviderClient:   taxClient,
		MarketplaceBilling:  nil,
		LicenseAuthority:    nil,
		LicenseCache:        licenseCache,
		OperationsMetrics:   opsmetrics.NewMetrics(),
		AdminConsoleMetrics: adminmetrics.NewMetrics(),
		IAMMode:             iamResolver.Mode(),
		IAMModeSource:       iamResolver.Source(),
		AuthProxy:           authClient,
		IAMDirectory:        localIAM,
		TaskBus:             taskBusClient,
	}

	wstransport.RegisterTaskBusBridge(deps.TaskBus)

	customerAuthenticator, localCustomerAuth, err := pluginbootstrap.BuildCustomerAuthenticator(cfg, deps)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize customer authenticator")
	}
	deps.CustomerAuthMode = cfg.ResolveCustomerAuthMode()
	deps.CustomerAuthenticator = customerAuthenticator
	deps.LocalCustomerAuth = localCustomerAuth

	listingRepo := marketplacerepo.NewListingRepository(queryDB)
	licenseRepoGlobal := marketplacerepo.NewLicenseRepository(queryDB)
	metricsProvider := recommendation.NewListingMetricsProvider(listingRepo)
	var syncJob *marketplacejobs.SyncJob
	if cfg == nil || cfg.Marketplace == nil || cfg.Marketplace.Recommendation.Enabled {
		syncJob = marketplacejobs.NewSyncJob(cfg, listingRepo, metricsProvider, logger.WithField("component", "marketplace_recommendation_sync"), listingRepo.ListTenantUuids)
	}

	var renewalJob *marketplacejobs.RenewalNotifier
	if cfg != nil && cfg.LicenseReminderLead() > 0 {
		renewalJob = marketplacejobs.NewLicenseRenewalNotifier(cfg, licenseRepoGlobal, logger.WithField("component", "marketplace_license_renewal_notifier"), listingRepo.ListTenantUuids, nil)
	}

	var credentialChecker *channelmasterjobs.CredentialChecker
	var metricRefresh *channelmasterjobs.MetricRefreshJob
	if deps.DB != nil {
		alertEmitter := channelobs.NewAlertEmitter(logger.WithField("component", "channel_master_alert"))
		lead := 7 * 24 * time.Hour
		interval := time.Hour
		credentialChecker = channelmasterjobs.NewCredentialChecker(deps, lead, interval, alertEmitter)
		metricRefresh = channelmasterjobs.NewMetricRefreshJob(deps, 30*time.Minute)
	}

	// 设置 gin engine 路由
	r := pluginrouter.NewRouter(cfg, deps)
	engine := r.Setup()

	// 创建 gRPC 服务器（可选）
	gs, err := grpcserver.NewGRPCServer(ctx, deps, cfg.GRPCServer)
	if err != nil {
		logger.WithError(err).Error("Failed to create gRPC server, continue with HTTP server only")
		gs = nil
	}

	appCfg := &fwbootstrap.Config{
		Listen:     cfg.Server.BindAddr,
		Env:        cfg.Server.Mode,
		Standalone: true,
		Gateway:    resolveFrameworkGatewayConfig(),
	}
	fwApp := fwbootstrap.NewApp(appCfg)

	if err := fwrouter.AttachHTTPServer(fwApp); err != nil {
		logger.WithError(err).Fatal("Failed to attach HTTP server")
	}
	fwrouter.RegisterFrameworkRoutes(fwApp)
	fwrouter.RegisterPluginRoutes(fwApp, func(r fwbootstrap.Router) {
		httpserver.RegisterGinRoutes(r, engine)
	})

	if err := manifest.Register(fwApp, manifestx.Plugin()); err != nil {
		logger.WithError(err).Fatal("Failed to register manifest")
	}

	// 使用 errgroup 并发启动服务器
	g, groupCtx := errgroup.WithContext(ctx)

	schedulerBridge := integrationjobs.NewSchedulerBridgeFromEnv(logger.WithField("component", "scheduler-bridge"))
	workerSpecs := make([]integrationjobs.WorkerSpec, 0, 4)
	if syncJob != nil {
		workerSpecs = append(workerSpecs, integrationjobs.WorkerSpec{
			Name:     syncJob.Name(),
			Interval: syncJob.Interval(),
			RunOnce:  syncJob.RunOnce,
		})
	}
	if renewalJob != nil {
		workerSpecs = append(workerSpecs, integrationjobs.WorkerSpec{
			Name:     renewalJob.Name(),
			Interval: renewalJob.Interval(),
			RunOnce:  renewalJob.RunOnce,
		})
	}
	if credentialChecker != nil {
		workerSpecs = append(workerSpecs, integrationjobs.WorkerSpec{
			Name:     credentialChecker.Name(),
			Interval: credentialChecker.Interval(),
			RunOnce:  credentialChecker.RunOnce,
		})
	}
	if metricRefresh != nil {
		workerSpecs = append(workerSpecs, integrationjobs.WorkerSpec{
			Name:     metricRefresh.Name(),
			Interval: metricRefresh.Interval(),
			RunOnce:  metricRefresh.RunOnce,
		})
	}
	g.Go(func() error {
		logger.WithField("addr", cfg.Server.BindAddr).Info("Starting HTTP server...")
		return fwApp.Run()
	})

	if gs != nil {
		g.Go(func() error {
			return gs.Serve(groupCtx)
		})
	}

	// 先启动 HTTP/GRPC，再并发注册调度任务，避免调度桥接阻塞健康检查。
	schedulerBridge.StartWorkers(groupCtx, g.Go, workerSpecs...)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在单独的 goroutine 中等待信号
	go func() {
		<-quit
		logger.Info("Shutting down servers...")

		cancel()

		if err := fwApp.Shutdown(); err != nil {
			logger.WithError(err).Error("HTTP server shutdown error")
		} else {
			logger.Info("HTTP server shutdown completed")
		}

		// 关闭数据库连接
		if err := dbpkg.Close(); err != nil {
			logger.WithError(err).Error("DB close error")
		} else {
			logger.Info("Database connection closed")
		}

		if gs != nil {
			gs.GracefulStop()
		}
	}()

	// 等待服务器启动失败或优雅关闭
	if err := g.Wait(); err != nil {
		logger.WithError(err).Error("Server error")
		os.Exit(1)
	}

	logger.Info("All servers shutdown completed")
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func resolveFrameworkGatewayConfig() fwbootstrap.GatewayConfig {
	toolToken, _ := pluginbootstrap.ResolveToolToken()
	return fwbootstrap.GatewayConfig{
		BaseURL:         resolveGatewayBaseURL(),
		ToolToken:       toolToken,
		TenantID:        strings.TrimSpace(os.Getenv("PX_TENANT_UUID")),
		GRPCTarget:      strings.TrimSpace(os.Getenv("PX_GATEWAY_GRPC_TARGET")),
		Timeout:         resolveGatewayTimeout(),
		UserAgent:       strings.TrimSpace(os.Getenv("PX_GATEWAY_USER_AGENT")),
		ContractVersion: strings.TrimSpace(os.Getenv("PX_GATEWAY_CONTRACT_VERSION")),
	}
}

func resolveGatewayBaseURL() string {
	base := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	return strings.TrimRight(base, "/")
}

func resolveGatewayAPIPrefix() string {
	prefix := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_PREFIX"))
	if prefix == "" {
		prefix = "/api/v1"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = "/" + strings.Trim(prefix, "/")
	if prefix == "/" {
		return ""
	}
	return prefix
}

func resolveGatewayTimeout() time.Duration {
	const fallback = 60 * time.Second
	raw := strings.TrimSpace(os.Getenv("PX_GATEWAY_TIMEOUT"))
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if sec, err := strconv.Atoi(raw); err == nil && sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return fallback
}

func normalizeGatewayAuthScheme(raw, toolToken, apiKey string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "apikey", "api-key", "api_key":
		return "apikey"
	case "bearer":
		return "bearer"
	}
	if strings.TrimSpace(apiKey) != "" {
		return "apikey"
	}
	return "bearer"
}
