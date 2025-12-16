package channel_master

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCredentialUpsertCreatesExpiringAlert(t *testing.T) {
	models.ForceSchemaForTests("")
	svc, ctx, db := newTestCredentialService(t)

	expires := time.Now().Add(48 * time.Hour).UTC()
	dto, err := svc.Upsert(ctx, "channel-expiring", CredentialUpsertInput{
		Type:          "oauth",
		Payload:       map[string]any{"token": "secret"},
		Scope:         []string{"read"},
		AttachmentURL: "https://files.example.com/contract.pdf",
		Metadata: map[string]any{
			"note": "unit",
		},
		ExpiresAt: &expires,
	})
	require.NoError(t, err)
	require.Equal(t, "expiring", dto.Status)

	var stored channelmodel.ChannelCredential
	require.NoError(t, db.WithContext(ctx).Where("channel_id = ?", "channel-expiring").First(&stored).Error)
	require.NotEmpty(t, stored.SecretCiphertext)
	require.NotEmpty(t, stored.SecretNonce)
	require.NotEmpty(t, stored.DekCiphertext)
	require.NotEmpty(t, stored.DekNonce)
	require.Equal(t, "https://files.example.com/contract.pdf", stored.AttachmentURL)
	require.NotNil(t, stored.LastRotatedAt)
	require.Equal(t, "static-v1", stored.KeyVersion)

	var alert channelmodel.ChannelAlert
	require.NoError(t, db.WithContext(ctx).
		Where("channel_id = ? AND type = ?", "channel-expiring", channelmodel.AlertTypeCredentialExpiring).
		First(&alert).Error)
	require.Equal(t, "open", alert.Status)
	require.Equal(t, "warning", alert.Severity)
}

func TestCredentialTestResultAlertsLifecycle(t *testing.T) {
	models.ForceSchemaForTests("")
	svc, ctx, db := newTestCredentialService(t)

	_, err := svc.Upsert(ctx, "channel-test", CredentialUpsertInput{
		Type:    "oauth",
		Payload: map[string]any{"token": "secret"},
	})
	require.NoError(t, err)

	_, err = svc.RecordTestResult(ctx, "channel-test", CredentialTestInput{
		Type:      "oauth",
		Succeeded: false,
		Result:    map[string]any{"code": 401},
	})
	require.NoError(t, err)

	var alert channelmodel.ChannelAlert
	require.NoError(t, db.WithContext(ctx).
		Where("channel_id = ? AND type = ?", "channel-test", channelmodel.AlertTypeCredentialTestFailed).
		First(&alert).Error)
	require.Equal(t, "open", alert.Status)

	_, err = svc.RecordTestResult(ctx, "channel-test", CredentialTestInput{
		Type:      "oauth",
		Succeeded: true,
	})
	require.NoError(t, err)

	require.NoError(t, db.WithContext(ctx).
		Where("channel_id = ? AND type = ?", "channel-test", channelmodel.AlertTypeCredentialTestFailed).
		First(&alert).Error)
	require.Equal(t, "resolved", alert.Status)
	require.NotNil(t, alert.ResolvedAt)
}

func newTestCredentialService(t *testing.T) (*CredentialService, context.Context, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	createCredentialTables(t, db)

	deps := &app.Deps{
		DB:  db,
		Ctx: context.Background(),
		Config: &config.Config{
			Server: &config.ServerConfig{SecretKey: "unit-test-secret"},
		},
	}
	svc := NewCredentialService(deps, nil, nil)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-cred-test")
	return svc, ctx, db
}

func createCredentialTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
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
			key_version TEXT,
			algorithm TEXT,
			expires_at DATETIME,
			last_refreshed_at DATETIME,
			last_tested_at DATETIME,
			last_rotated_at DATETIME,
			test_result TEXT,
			metadata TEXT,
			attachment_url TEXT,
			created_by TEXT,
			updated_by TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_cred_tenant_type ON channel_credentials(tenant_uuid, channel_id, type);`,
		`CREATE TABLE IF NOT EXISTS channel_alerts (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			type TEXT NOT NULL,
			severity TEXT,
			title TEXT,
			description TEXT,
			status TEXT NOT NULL,
			assignee_uuid TEXT,
			task_id TEXT,
			triggered_at DATETIME NOT NULL,
			resolved_at DATETIME,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
