package performance

import (
	"context"
	"sort"
	"testing"
	"time"
)

func TestPromotionQuoteSLA(t *testing.T) {
	ctx := context.Background()
	svc := buildPromotionQuoteFixture(t)
	input := promotionQuoteInput()

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
		t.Fatalf("promotion quote SLA breached: p95=%s > 500ms", p95)
	}
}
