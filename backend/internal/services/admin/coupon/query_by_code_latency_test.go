package coupon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryLatency_ByCouponCode_P95Under500ms(t *testing.T) {
	ctx := context.Background()
	svc := buildQueryLatencyFixture(t, 3000)

	const rounds = 80
	durations := make([]time.Duration, 0, rounds)
	for i := 0; i < rounds; i++ {
		start := time.Now()
		res, err := svc.ListUsageLogs(ctx, "tenant-1", UsageLogQueryFilter{CouponCode: "CPN-000", Page: 1, PageSize: 20})
		require.NoError(t, err)
		require.NotNil(t, res)
		durations = append(durations, time.Since(start))
	}
	p95 := percentile95(durations)
	require.LessOrEqual(t, p95, 500*time.Millisecond, "coupon-code query p95 too high: %s", p95)
}
