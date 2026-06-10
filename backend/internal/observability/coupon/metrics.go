package coupon

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

const (
	metricCouponIssuedTotal     = "powerx_coupon_issued_total"
	metricCouponReservedTotal   = "powerx_coupon_reserved_total"
	metricCouponRedeemedTotal   = "powerx_coupon_redeemed_total"
	metricCouponReleasedTotal   = "powerx_coupon_released_total"
	metricCouponRefundedTotal   = "powerx_coupon_refunded_total"
	metricCouponActionFailTotal = "powerx_coupon_action_failed_total"
)

// Metrics 提供优惠券动作计数器（发放/预占/核销/释放/返券）。
type Metrics struct {
	mu       sync.RWMutex
	counters map[string]map[string]float64
}

func NewMetrics() *Metrics {
	return &Metrics{counters: map[string]map[string]float64{}}
}

func (m *Metrics) RecordIssued(channel, result string) {
	m.inc(metricCouponIssuedTotal, channel, result)
}
func (m *Metrics) RecordReserved(channel, result string) {
	m.inc(metricCouponReservedTotal, channel, result)
}
func (m *Metrics) RecordRedeemed(channel, result string) {
	m.inc(metricCouponRedeemedTotal, channel, result)
}
func (m *Metrics) RecordReleased(channel, result string) {
	m.inc(metricCouponReleasedTotal, channel, result)
}
func (m *Metrics) RecordRefunded(channel, result string) {
	m.inc(metricCouponRefundedTotal, channel, result)
}

func (m *Metrics) RecordActionFailed(action, reason string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	labels := labelKey(map[string]string{
		"action": normalize(action),
		"reason": normalize(reason),
	})
	ensureCounter(m.counters, metricCouponActionFailTotal)[labels]++
}

func (m *Metrics) inc(metric, channel, result string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	labels := labelKey(map[string]string{
		"channel": normalize(channel),
		"result":  normalize(result),
	})
	ensureCounter(m.counters, metric)[labels]++
}

func (m *Metrics) RenderPrometheus(w io.Writer) {
	if m == nil {
		return
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for metric, series := range m.counters {
		fmt.Fprintf(w, "# TYPE %s counter\n", metric)
		for _, labels := range sortedKeys(series) {
			fmt.Fprintf(w, "%s{%s} %g\n", metric, labels, series[labels])
		}
	}
}

func ensureCounter(store map[string]map[string]float64, metric string) map[string]float64 {
	if store[metric] == nil {
		store[metric] = make(map[string]float64)
	}
	return store[metric]
}

func labelKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = fmt.Sprintf("%s=\"%s\"", k, labels[k])
	}
	return strings.Join(pairs, ",")
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func normalize(s string) string {
	v := strings.TrimSpace(strings.ToLower(s))
	if v == "" {
		return "unknown"
	}
	return v
}
