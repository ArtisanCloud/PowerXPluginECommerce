package channel_master

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStrategyService_UpsertStrategy(t *testing.T) {
	svc, ctx, db := newTestStrategyService(t)
	channelID := "chan-001"
	now := time.Now().UTC().Truncate(time.Second)
	insertChannel(t, db, channelID, "tenant-strategy", "owner-initial", nil)

	fee := 12.5
	payload := StrategyUpsertInput{
		Strategy: ChannelStrategyInput{
			PricebookID:     "pricebook-1",
			FeeRate:         &fee,
			SettlementCycle: "net30",
			PaymentTerms:    "预付30%",
			Notes:           "测试策略",
			EffectiveAt:     &now,
		},
		Team: ChannelTeamInput{
			OwnerUUID:    "owner-updated",
			ApproverUUID: "approver-1",
			Operators:    []string{"ops-a", "ops-b"},
		},
	}
	snapshot, err := svc.UpsertStrategy(ctx, channelID, payload)
	require.NoError(t, err)
	require.Equal(t, "owner-updated", snapshot.Team.OwnerUUID)
	require.Equal(t, "approver-1", snapshot.Team.ApproverUUID)
	require.Equal(t, 2, len(snapshot.Team.Operators))
	require.Equal(t, "pricebook-1", snapshot.Config.PricebookID)
	require.Equal(t, 12.5, snapshot.Config.FeeRate)

	got, err := svc.GetStrategy(ctx, channelID)
	require.NoError(t, err)
	require.Equal(t, snapshot.Team.OwnerUUID, got.Team.OwnerUUID)
	require.Equal(t, snapshot.Config.PricebookID, got.Config.PricebookID)
}

func insertChannel(t *testing.T, db *gorm.DB, channelID, tenant, owner string, approver *string) {
	t.Helper()
	stmt := `INSERT INTO channel_masters (id, tenant_uuid, owner_uuid, approver_uuid, platform, channel_type, store_id, name, region, status, tags, created_at, updated_at)
	VALUES (?, ?, ?, ?, 'tmall', 'platform_manual', ?, ?, 'CN', 'authorized', '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);`
	require.NoError(t, db.Exec(stmt, channelID, tenant, owner, approver, channelID, "Channel "+channelID).Error)
}

func newTestStrategyService(t *testing.T) (*StrategyService, context.Context, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS channel_masters (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT,
			owner_uuid TEXT,
			approver_uuid TEXT,
			platform TEXT,
			channel_type TEXT,
			store_id TEXT,
			name TEXT,
			region TEXT,
			status TEXT,
			tags TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS channel_configs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT,
			channel_id TEXT,
			pricebook_id TEXT,
			inventory_strategy_id TEXT,
			logistics_strategy_id TEXT,
			cs_sla_id TEXT,
			fee_rate REAL,
			settlement_cycle TEXT,
			payment_terms TEXT,
			notes TEXT,
			effective_at DATETIME,
			deprecated_at DATETIME,
			team_metadata JSON,
			created_by TEXT,
			updated_by TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_config_tenant_channel ON channel_configs(tenant_uuid, channel_id);`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
	deps := &app.Deps{
		DB: db,
		Config: &config.Config{
			Server: &config.ServerConfig{SecretKey: "strategy-test"},
		},
	}
	audit := channelobs.NewAuditEmitter(nil)
	svc := NewStrategyService(deps, audit)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-strategy")
	return svc, ctx, db
}
