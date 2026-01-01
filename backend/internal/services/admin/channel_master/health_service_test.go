package channel_master

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMetricsHealthScore_GatesOnCredentials(t *testing.T) {
	svc, ctx, db := newTestMetricsService(t)
	ts := time.Now()
	require.NoError(t, db.WithContext(ctx).Exec(`INSERT INTO channel_metrics
		(id, tenant_uuid, channel_id, window, gmv, orders, gmv_growth_rate, inventory_coverage, error_rate, sync_success_rate, health_score, source_timestamp)
		VALUES ('metric-1', 'tenant-health', 'channel-a', 'd7', 1000, 20, 0.05, 0.90, 0.01, 0.99, 80, ?)`, ts).Error)
	require.NoError(t, db.WithContext(ctx).Exec(`INSERT INTO channel_credentials
		(id, tenant_uuid, channel_id, type, status, scope, secret_ciphertext, secret_nonce, dek_ciphertext, dek_nonce, algorithm, key_version)
		VALUES ('cred-1', 'tenant-health', 'channel-a', 'oauth', 'expired', '{}', x'01', x'01', x'02', x'02', 'AES-GCM', 'static-v1')`).Error)

	metrics, err := svc.LoadMetrics(ctx, "channel-a")
	require.NoError(t, err)
	require.Len(t, metrics, 1)

	health := svc.ComputeHealthScore(ctx, "channel-a", metrics)
	require.NotEqual(t, 0, health.Score)
	require.Contains(t, health.Labels, "credential_expired")
	require.Less(t, health.Score, 80)
}

func TestMetricsHealthScore_FlagsErrorRate(t *testing.T) {
	svc, ctx, db := newTestMetricsService(t)
	ts := time.Now()
	require.NoError(t, db.WithContext(ctx).Exec(`INSERT INTO channel_metrics
		(id, tenant_uuid, channel_id, window, gmv, orders, gmv_growth_rate, inventory_coverage, error_rate, sync_success_rate, health_score, source_timestamp)
		VALUES ('metric-2', 'tenant-health', 'channel-b', 'd7', 1000, 20, -0.10, 0.70, 0.3, 0.80, 60, ?)`, ts).Error)

	// Without metrics slice defaults to placeholder; pass loaded metrics to exercise logic.
	metrics, err := svc.LoadMetrics(ctx, "channel-b")
	require.NoError(t, err)
	health := svc.ComputeHealthScore(ctx, "channel-b", metrics)
	require.Contains(t, health.Labels, "error_rate_high")
	require.Contains(t, health.Labels, "sync_unstable")
	require.Less(t, health.Score, 60)
}

func newTestMetricsService(t *testing.T) (*MetricsService, context.Context, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS channel_metrics (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			window TEXT,
			gmv REAL,
			orders INTEGER,
			gmv_growth_rate REAL,
			inventory_coverage REAL,
			error_rate REAL,
			sync_success_rate REAL,
			health_score INTEGER,
			source_timestamp DATETIME,
			source_job_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS channel_credentials (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			scope TEXT,
			secret_ciphertext BLOB NOT NULL,
			secret_nonce BLOB NOT NULL,
			dek_ciphertext BLOB NOT NULL,
			dek_nonce BLOB NOT NULL,
			algorithm TEXT,
			key_version TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}

	deps := &app.Deps{
		DB: db,
		Config: &config.Config{
			Server: &config.ServerConfig{SecretKey: "metrics-test"},
		},
	}
	svc := NewMetricsService(deps, nil)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-health")
	return svc, ctx, db
}
