package subscription_reconciliation

import (
	"context"
	"fmt"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestReconciliationPerformance_10kBenchmark(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_perf_10k")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "f7f4b140-01e3-4e8a-b77a-0ed2e85702a8"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	samples := make([]ReconciliationSample, 0, 10000)
	for i := 0; i < 10000; i++ {
		amount := int64(100 + i%13)
		samples = append(samples, ReconciliationSample{
			SubscriptionRef:     fmt.Sprintf("sub-%05d", i),
			BillRef:             fmt.Sprintf("bill-%05d", i),
			ExpectedAmountMinor: amount,
			ActualAmountMinor:   amount,
			ReasonCode:          "perf_case",
		})
	}

	start := time.Now()
	_, err := svc.CreateBatch(ctx, tenantUUID, CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		CreatedBy:    "perf",
		Samples:      samples,
	})
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Less(t, elapsed, 30*time.Minute, "10k 对账编排应在 30 分钟内完成")
	t.Logf("10k reconciliation elapsed=%s", elapsed)
}
