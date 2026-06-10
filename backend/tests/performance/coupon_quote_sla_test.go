package performance

import (
	"context"
	"sort"
	"testing"
	"time"

	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
)

// TestCouponQuoteSLA 校验试算接口在本地基准数据下 P95 <= 500ms。
func TestCouponQuoteSLA(t *testing.T) {
	ctx := context.Background()
	svc := buildQuoteServiceFixture(t)
	now := time.Now().UTC()
	input := couponsvc.QuoteInput{
		TenantUUID: "tenant-1",
		UserID:     "user-1",
		Channel:    "official",
		CouponIDs:  []string{"asset-1", "asset-2"},
		Items: []couponsvc.QuoteItemInput{
			{LineID: "l1", SKUID: "sku-1", Qty: 1, UnitPriceMinor: 1200},
			{LineID: "l2", SKUID: "sku-2", Qty: 1, UnitPriceMinor: 800},
		},
		Now:      now,
		Currency: "CNY",
	}

	const rounds = 120
	durations := make([]time.Duration, 0, rounds)
	for i := 0; i < rounds; i++ {
		start := time.Now()
		if _, err := svc.Quote(ctx, input); err != nil {
			t.Fatalf("quote failed: %v", err)
		}
		durations = append(durations, time.Since(start))
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[int(float64(len(durations))*0.95)-1]
	if p95 > 500*time.Millisecond {
		t.Fatalf("quote SLA breached: p95=%s > 500ms", p95)
	}
}
