package order

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrderRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:order_repo_tests?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		order_no TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		channel TEXT NOT NULL,
		status TEXT NOT NULL,
		currency TEXT NOT NULL,
		subtotal_amount BIGINT NOT NULL DEFAULT 0,
		total_amount BIGINT NOT NULL DEFAULT 0,
		shipping_address_id TEXT,
		shipping_address_snapshot TEXT,
		price_snapshot TEXT,
		sellability_snapshot TEXT,
		created_by_type TEXT,
		created_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);`).Error)
	return db
}

func TestOrderRepository_List_FilterByCustomerAndStatus(t *testing.T) {
	db := setupOrderRepoTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	const tenant = "tenant-1"
	now := time.Now().UTC()

	seed := []ordermodel.Order{
		{ID: "o1", TenantUUID: tenant, OrderNo: "O1", CustomerID: "c1", Channel: "official", Status: "pending_payment", Currency: "CNY", CreatedAt: now.Add(-3 * time.Minute)},
		{ID: "o2", TenantUUID: tenant, OrderNo: "O2", CustomerID: "c1", Channel: "official", Status: "cancelled", Currency: "CNY", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "o3", TenantUUID: tenant, OrderNo: "O3", CustomerID: "c2", Channel: "official", Status: "pending_payment", Currency: "CNY", CreatedAt: now.Add(-1 * time.Minute)},
	}
	require.NoError(t, db.WithContext(ctx).Create(&seed).Error)

	got, err := repo.List(ctx, tenant, OrderListFilter{CustomerID: "c1", Status: "pending_payment"}, 1, 20)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, int64(1), got.Total)
	require.Len(t, got.Items, 1)
	require.Equal(t, "o1", got.Items[0].ID)
}
