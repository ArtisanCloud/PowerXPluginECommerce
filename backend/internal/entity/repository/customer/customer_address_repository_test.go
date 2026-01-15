package customer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestCustomerAddressRepository_GetByID_ScopedByCustomer(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
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

	repo := NewCustomerAddressRepository(db)
	const tenant = "tenant-test"
	const customerA = "cust-a"
	const customerB = "cust-b"

	require.NoError(t, db.Exec(
		`INSERT INTO customer_addresses (id, tenant_uuid, customer_id, is_default, recipient_name, recipient_phone, address1, created_at, updated_at)
		 VALUES (?, ?, ?, 1, ?, ?, ?, datetime('now'), datetime('now'))`,
		"addr-1", tenant, customerA, "张三", "13800138000", "北京市朝阳区",
	).Error)

	_, err := repo.GetByID(ctx, tenant, customerB, "addr-1")
	require.Error(t, err)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
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
