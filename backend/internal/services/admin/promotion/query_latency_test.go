package promotion

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPromotionQueryLatency_ByOrderID_P95Under500ms(t *testing.T) {
	ctx := context.Background()
	svc := buildPromotionQueryLatencyFixture(t, 3000)

	const rounds = 80
	durations := make([]time.Duration, 0, rounds)
	for i := 0; i < rounds; i++ {
		start := time.Now()
		row, err := svc.FindByOrderID(ctx, "tenant-1", "ord-777")
		if err != nil {
			t.Fatalf("find snapshot: %v", err)
		}
		if row.OrderID != "ord-777" {
			t.Fatalf("unexpected order id %q", row.OrderID)
		}
		durations = append(durations, time.Since(start))
	}
	if p95 := promotionPercentile95(durations); p95 > 500*time.Millisecond {
		t.Fatalf("promotion order lookup p95 too high: %s", p95)
	}
}

func buildPromotionQueryLatencyFixture(t *testing.T, rows int) *SnapshotService {
	t.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:promotion_query_latency_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS order_promotion_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor INTEGER NOT NULL,
			promotion_discount_minor INTEGER NOT NULL,
			after_promotion_total_minor INTEGER NOT NULL,
			applied_promotions TEXT NOT NULL,
			rejected_promotions TEXT NOT NULL,
			line_allocations TEXT NOT NULL,
			priced_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_promotion_snapshot_tenant_order ON order_promotion_snapshots(tenant_uuid, order_id)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	now := time.Now().UTC()
	for i := 0; i < rows; i++ {
		orderID := fmt.Sprintf("ord-%d", i)
		if i == rows/2 {
			orderID = "ord-777"
		}
		if err := db.Exec(`INSERT INTO order_promotion_snapshots
			(id, tenant_uuid, order_id, currency, base_total_minor, promotion_discount_minor, after_promotion_total_minor, applied_promotions, rejected_promotions, line_allocations, priced_at, created_at, updated_at)
			VALUES (?, 'tenant-1', ?, 'CNY', 1000, 100, 900, '[]', '[]', '[]', ?, ?, ?)`,
			fmt.Sprintf("snap-%d", i), orderID, now, now, now).Error; err != nil {
			t.Fatalf("seed snapshot: %v", err)
		}
	}
	return NewSnapshotService(&app.Deps{DB: db})
}

func promotionPercentile95(samples []time.Duration) time.Duration {
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
