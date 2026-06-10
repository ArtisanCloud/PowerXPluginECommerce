package coupon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQueryLatency_ByOrderID_P95Under500ms(t *testing.T) {
	ctx := context.Background()
	svc := buildQueryLatencyFixture(t, 3000)

	const rounds = 80
	durations := make([]time.Duration, 0, rounds)
	for i := 0; i < rounds; i++ {
		start := time.Now()
		res, err := svc.ListUsageLogs(ctx, "tenant-1", UsageLogQueryFilter{OrderID: "ord-777", Page: 1, PageSize: 20})
		require.NoError(t, err)
		require.NotNil(t, res)
		durations = append(durations, time.Since(start))
	}
	p95 := percentile95(durations)
	require.LessOrEqual(t, p95, 500*time.Millisecond, "order-id query p95 too high: %s", p95)
}

func buildQueryLatencyFixture(t *testing.T, rows int) *QueryService {
	t.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:coupon_query_latency_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS coupon_assets (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			template_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			coupon_code TEXT NOT NULL,
			status TEXT NOT NULL,
			reserved_order_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_usage_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			request_id TEXT,
			created_by TEXT,
			created_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_tenant_order ON coupon_usage_logs(tenant_uuid, order_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_tenant_code ON coupon_assets(tenant_uuid, coupon_code)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}

	now := time.Now().UTC()
	for i := 0; i < rows; i++ {
		assetID := fmt.Sprintf("asset-%d", i)
		orderID := fmt.Sprintf("ord-%d", i%1000)
		if i%19 == 0 {
			orderID = "ord-777"
		}
		code := fmt.Sprintf("CPN-%06d", i)
		require.NoError(t, db.Exec(`INSERT INTO coupon_assets (id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			assetID, "tenant-1", "tpl-1", fmt.Sprintf("u-%d", i%200), code, "redeemed", orderID, now, now).Error)
		require.NoError(t, db.Exec(`INSERT INTO coupon_usage_logs (id, tenant_uuid, asset_id, order_id, action, action_reason, idempotency_key, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("log-%d", i), "tenant-1", assetID, orderID, "redeem", "payment_success", fmt.Sprintf("redeem:%s:%s", orderID, assetID), now).Error)
	}
	return NewQueryService(&app.Deps{DB: db})
}

func percentile95(samples []time.Duration) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	dup := append([]time.Duration(nil), samples...)
	for i := 0; i < len(dup)-1; i++ {
		for j := i + 1; j < len(dup); j++ {
			if dup[j] < dup[i] {
				dup[i], dup[j] = dup[j], dup[i]
			}
		}
	}
	idx := int(float64(len(dup))*0.95) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(dup) {
		idx = len(dup) - 1
	}
	return dup[idx]
}
