package membership

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceListTokenTransactions_FilterBySourceIDFuzzy(t *testing.T) {
	ctx := context.Background()
	svc, db := setupMembershipServiceTest(t)

	const tenant = "tenant-1"
	const customer = "customer-1"
	now := time.Now().UTC()

	rows := []membershipModel.TokenTransaction{
		{
			ID:         "tx-1",
			TenantUUID: tenant,
			CustomerID: customer,
			TokenCode:  "POINTS",
			Delta:      10,
			SourceType: "mall_redeem",
			SourceID:   "media-assets-001",
			CreatedAt:  now.Add(-2 * time.Minute),
		},
		{
			ID:         "tx-2",
			TenantUUID: tenant,
			CustomerID: customer,
			TokenCode:  "POINTS",
			Delta:      12,
			SourceType: "mall_redeem",
			SourceID:   "assets-order-002",
			CreatedAt:  now.Add(-1 * time.Minute),
		},
		{
			ID:         "tx-3",
			TenantUUID: tenant,
			CustomerID: customer,
			TokenCode:  "POINTS",
			Delta:      15,
			SourceType: "mall_redeem",
			SourceID:   "coupon-003",
			CreatedAt:  now,
		},
	}
	require.NoError(t, db.WithContext(ctx).Create(&rows).Error)

	items, total, err := svc.ListTokenTransactions(ctx, tenant, ListTokenTransactionsRequest{
		CustomerID: customer,
		SourceID:   "assets",
		Page:       1,
		PageSize:   20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)
	for _, item := range items {
		require.Contains(t, item.SourceID, "assets")
	}
}

func TestServiceListTokenTransactions_FilterBySourceIDEscapedLiteral(t *testing.T) {
	ctx := context.Background()
	svc, db := setupMembershipServiceTest(t)

	const tenant = "tenant-1"
	const customer = "customer-1"
	now := time.Now().UTC()

	rows := []membershipModel.TokenTransaction{
		{
			ID:         "tx-1",
			TenantUUID: tenant,
			CustomerID: customer,
			TokenCode:  "POINTS",
			Delta:      10,
			SourceType: "mall_redeem",
			SourceID:   `order_%\_001`,
			CreatedAt:  now.Add(-2 * time.Minute),
		},
		{
			ID:         "tx-2",
			TenantUUID: tenant,
			CustomerID: customer,
			TokenCode:  "POINTS",
			Delta:      10,
			SourceType: "mall_redeem",
			SourceID:   "order-a-b-c",
			CreatedAt:  now.Add(-1 * time.Minute),
		},
	}
	require.NoError(t, db.WithContext(ctx).Create(&rows).Error)

	items, total, err := svc.ListTokenTransactions(ctx, tenant, ListTokenTransactionsRequest{
		CustomerID: customer,
		SourceID:   `_%\_`,
		Page:       1,
		PageSize:   20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, `order_%\_001`, items[0].SourceID)
}

func setupMembershipServiceTest(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	coremodels.ForceSchemaForTests("")

	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS token_transactions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		token_code TEXT NOT NULL,
		delta INTEGER NOT NULL,
		source_type TEXT NOT NULL,
		source_id TEXT NOT NULL,
		created_at DATETIME
	)`).Error)

	return NewService(&app.Deps{DB: db}), db
}
