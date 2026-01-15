package customer_address

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestService_DefaultUniqueness(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createTables(t, db)

	const tenant = "tenant-test"
	const customerID = "cust-1"

	svc := NewService(&app.Deps{DB: db})

	first, err := svc.CreateAddress(ctx, tenant, customerID, AddressUpsertRequest{
		ShippingAddress: ShippingAddress{
			RecipientName:  "张三",
			RecipientPhone: "13800138000",
			Address1:       "北京市朝阳区",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, first)
	require.True(t, first.IsDefault)

	makeDefault := true
	second, err := svc.CreateAddress(ctx, tenant, customerID, AddressUpsertRequest{
		IsDefault: &makeDefault,
		ShippingAddress: ShippingAddress{
			RecipientName:  "李四",
			RecipientPhone: "13800138001",
			Address1:       "上海市浦东新区",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, second)
	require.True(t, second.IsDefault)

	var defaultCount int64
	require.NoError(t, db.Raw(
		`SELECT COUNT(1) FROM customer_addresses WHERE tenant_uuid = ? AND customer_id = ? AND is_default = 1`,
		tenant, customerID,
	).Scan(&defaultCount).Error)
	require.Equal(t, int64(1), defaultCount)

	list, err := svc.ListAddresses(ctx, tenant, customerID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, second.ID, list[0].ID)
	require.True(t, list[0].IsDefault)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	return db
}

func createTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(
		`CREATE TABLE IF NOT EXISTS customer_addresses (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			customer_id TEXT NOT NULL,
			is_default BOOLEAN NOT NULL DEFAULT 0,
			label TEXT,
			recipient_name TEXT NOT NULL,
			recipient_phone TEXT NOT NULL,
			country_code TEXT,
			province TEXT,
			city TEXT,
			district TEXT,
			address1 TEXT NOT NULL,
			address2 TEXT,
			postal_code TEXT,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	).Error)
}
