package promotion

import "sync/atomic"

// Metrics keeps lightweight in-process counters for promotion quote visibility.
type Metrics struct {
	quoteTotal      atomic.Int64
	quoteErrors     atomic.Int64
	discountMinor   atomic.Int64
	rejectedTotal    atomic.Int64
}

func (m *Metrics) RecordQuote(discountMinor int64, rejected int, failed bool) {
	if m == nil {
		return
	}
	m.quoteTotal.Add(1)
	if failed {
		m.quoteErrors.Add(1)
	}
	if discountMinor > 0 {
		m.discountMinor.Add(discountMinor)
	}
	if rejected > 0 {
		m.rejectedTotal.Add(int64(rejected))
	}
}

func (m *Metrics) Snapshot() map[string]int64 {
	if m == nil {
		return map[string]int64{}
	}
	return map[string]int64{
		"quote_total":          m.quoteTotal.Load(),
		"quote_errors":         m.quoteErrors.Load(),
		"discount_minor_total": m.discountMinor.Load(),
		"rejected_total":       m.rejectedTotal.Load(),
	}
}
