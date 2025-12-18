package channel_master

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestChannelLifecycleCreateSubmitApprove(t *testing.T) {
	models.ForceSchemaForTests("")
	db := newTestDB(t)
	createChannelTable(t, db)
	deps := &app.Deps{DB: db}
	svc := NewService(deps, nil, &stubAudit{}, nil)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-test")
	ctx = context.WithValue(ctx, "tenant_ctx", authx.TenantContext{TenantUUID: "tenant-test", UserID: 42})

	input := CreateChannelInput{
		Name:         "旗舰店",
		Platform:     "tmall",
		StoreID:      "tmall-001",
		Region:       "CN",
		OwnerUUID:    "owner-1",
		ApproverUUID: "approver-1",
		ChannelType:  ChannelTypePlatformOAuth,
		Domain:       "store.example.com",
		Tags:         []string{"data_gap", "alert"},
		Contact: ChannelContactInput{
			Name:  "张三",
			Phone: "123456",
			Email: "ops@example.com",
		},
	}
	created, err := svc.CreateDraft(ctx, input)
	require.NoError(t, err)
	require.Equal(t, StatusDraft, created.Status)

	// duplicate store detection
	_, err = svc.CreateDraft(ctx, input)
	require.ErrorIs(t, err, ErrDuplicateStore)

	input.Name = "旗舰店-更新"
	updated, err := svc.UpdateDraft(ctx, created.ID, input)
	require.NoError(t, err)
	require.Equal(t, "旗舰店-更新", updated.Name)

	submitted, err := svc.Submit(ctx, created.ID, SubmitChannelInput{Note: "ready"})
	require.NoError(t, err)
	require.Equal(t, StatusPendingReview, submitted.Status)

	approved, err := svc.ProcessApproval(ctx, created.ID, ApprovalDecisionInput{Decision: "approve"})
	require.NoError(t, err)
	require.Equal(t, StatusUnauthorized, approved.Status)

	// once approved, submit again should fail
	_, err = svc.Submit(ctx, created.ID, SubmitChannelInput{})
	require.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestSubmitOfflineRequiresEvidence(t *testing.T) {
	models.ForceSchemaForTests("")
	db := newTestDB(t)
	createChannelTable(t, db)
	deps := &app.Deps{DB: db}
	svc := NewService(deps, nil, &stubAudit{}, nil)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-test")

	input := CreateChannelInput{
		Name:        "线下店",
		Platform:    "offline",
		StoreID:     "offline-1",
		Region:      "CN",
		OwnerUUID:   "owner-9",
		ChannelType: ChannelTypeOffline,
		Contact: ChannelContactInput{
			Name:  "李四",
			Phone: "9876",
			Email: "offline@example.com",
		},
	}
	created, err := svc.CreateDraft(ctx, input)
	require.NoError(t, err)

	_, err = svc.Submit(ctx, created.ID, SubmitChannelInput{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "offlineEvidenceUrl")

	_, err = svc.Submit(ctx, created.ID, SubmitChannelInput{OfflineEvidenceURL: "https://example.com/evidence.pdf"})
	require.NoError(t, err)
}

type stubAudit struct{}

func (s *stubAudit) EmitChannelAudit(ctx context.Context, action string, payload map[string]any) error {
	return nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

func createChannelTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS channel_masters (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			platform TEXT NOT NULL,
			channel_type TEXT NOT NULL,
			store_id TEXT NOT NULL,
			name TEXT NOT NULL,
			domain TEXT,
			region TEXT NOT NULL,
			status TEXT NOT NULL,
			tags TEXT,
			owner_uuid TEXT NOT NULL,
			approver_uuid TEXT,
			contact_name TEXT,
			contact_phone TEXT,
			contact_email TEXT,
			health_score INTEGER,
			last_sync_at DATETIME,
			sync_status TEXT,
			created_by TEXT,
			updated_by TEXT,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_master_tenant_store ON channel_masters(tenant_uuid, store_id);`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
