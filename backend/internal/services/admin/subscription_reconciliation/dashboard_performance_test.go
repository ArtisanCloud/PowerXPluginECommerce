package subscription_reconciliation

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestDashboardPerformance_P95LessThan1s(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_dashboard_perf")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "5a1d93c2-df1e-4f82-b427-102f3882d3a7"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_batches
		(id, tenant_uuid, billing_cycle, run_type, expected_amount_minor, actual_amount_minor, delta_amount_minor, delta_count, status, started_at, finished_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"batch-dashboard-perf", tenantUUID, "2026-03-31", "daily", 100000, 90000, 10000, 300, "completed", now, now, now, now,
	).Error)

	for i := 0; i < 1200; i++ {
		deltaType := "missing_payment"
		if i%3 == 1 {
			deltaType = "amount_mismatch"
		}
		if i%3 == 2 {
			deltaType = "status_mismatch"
		}
		require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas
			(id, tenant_uuid, batch_id, subscription_ref, delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor, reason_code, status, delta_fingerprint, detected_at, metadata, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("delta-dashboard-perf-%d", i), tenantUUID, "batch-dashboard-perf", fmt.Sprintf("sub-%d", i), deltaType, "medium", 100, 50, 50, "always_fail", "open", fmt.Sprintf("fp-%d", i), now, `{"channel":"app","plan":"pro","region":"CN"}`, now, now,
		).Error)
	}

	durations := make([]time.Duration, 0, 40)
	for i := 0; i < 40; i++ {
		start := time.Now()
		_, err := svc.Dashboard(ctx, tenantUUID, DashboardQuery{
			From:          "2026-03-01",
			To:            "2026-03-31",
			Channel:       "app",
			Plan:          "pro",
			Region:        "CN",
			FailureReason: "always_fail",
		})
		require.NoError(t, err)
		durations = append(durations, time.Since(start))
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[int(float64(len(durations))*0.95)-1]
	require.Less(t, p95, time.Second, "dashboard 查询 p95 应低于 1 秒")
	t.Logf("dashboard p95=%s", p95)
}
