package subscription_reconciliation

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

const (
	metricDeltaCaseTotal           = "subscription_reconciliation_delta_cases_total"
	metricGovernanceRecoveredTotal = "subscription_reconciliation_governance_recovered_total"
	metricGovernanceFailedTotal    = "subscription_reconciliation_governance_failed_total"
)

var (
	kpiMu       sync.RWMutex
	kpiCounters = map[string]map[string]float64{}
)

func ensureKPICounter(metric string) map[string]float64 {
	if kpiCounters[metric] == nil {
		kpiCounters[metric] = map[string]float64{}
	}
	return kpiCounters[metric]
}

func kpiLabelKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
	}
	return strings.Join(parts, ",")
}

func recordKPI(metric string, labels map[string]string, delta float64) {
	kpiMu.Lock()
	defer kpiMu.Unlock()
	ensureKPICounter(metric)[kpiLabelKey(labels)] += delta
}

// RecordDeltaCase tracks generated reconciliation deltas.
func RecordDeltaCase(tenantUUID, deltaType string) {
	recordKPI(metricDeltaCaseTotal, map[string]string{
		"tenant_uuid": tenantUUID,
		"delta_type":  deltaType,
	}, 1)
}

// RecordGovernanceRecovery tracks recovered subscriptions from retry governance.
func RecordGovernanceRecovery(tenantUUID string) {
	recordKPI(metricGovernanceRecoveredTotal, map[string]string{
		"tenant_uuid": tenantUUID,
	}, 1)
}

// RecordGovernanceFailure tracks subscriptions still failed after governance run.
func RecordGovernanceFailure(tenantUUID string) {
	recordKPI(metricGovernanceFailedTotal, map[string]string{
		"tenant_uuid": tenantUUID,
	}, 1)
}

// RenderKPIMetrics writes KPI counters in Prometheus exposition format.
func RenderKPIMetrics(w io.Writer) {
	kpiMu.RLock()
	defer kpiMu.RUnlock()
	for metric, values := range kpiCounters {
		_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", metric)
		for label, value := range values {
			if label == "" {
				_, _ = fmt.Fprintf(w, "%s %v\n", metric, value)
				continue
			}
			_, _ = fmt.Fprintf(w, "%s{%s} %v\n", metric, label, value)
		}
	}
}

// ResetKPIMetrics clears all reconciliation KPI counters.
func ResetKPIMetrics() {
	kpiMu.Lock()
	defer kpiMu.Unlock()
	kpiCounters = map[string]map[string]float64{}
}
