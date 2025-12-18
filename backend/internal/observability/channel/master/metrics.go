package master

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Metrics aggregates lightweight counters for channel flows.
type Metrics struct {
	logger *logrus.Entry

	approvalLeadMs atomic.Uint64
	approvalCount  atomic.Uint64

	syncSuccess atomic.Uint64
	syncTotal   atomic.Uint64

	credentialValid atomic.Uint64
	credentialTotal atomic.Uint64

	kpiBuckets [4]atomic.Uint64
}

// MetricSnapshot exposes averages for dashboards.
type MetricSnapshot struct {
	AverageApprovalLead time.Duration    `json:"averageApprovalLead"`
	SyncSuccessRate     float64          `json:"syncSuccessRate"`
	CredentialCoverage  float64          `json:"credentialCoverage"`
	KPILoadHistogram    KPILoadHistogram `json:"kpiLoadHistogram"`
}

// KPILoadHistogram captures KPI load distribution buckets.
type KPILoadHistogram struct {
	Under1s uint64 `json:"under_1s"`
	Under2s uint64 `json:"under_2s"`
	Under3s uint64 `json:"under_3s"`
	Over3s  uint64 `json:"over_3s"`
}

// NewMetrics constructs a collector with optional logger enrichment.
func NewMetrics(logger *logrus.Entry) *Metrics {
	if logger == nil {
		logger = logrus.New().WithField("component", "channel-master-metrics")
	}
	return &Metrics{logger: logger}
}

// ObserveApprovalLead tracks durations between submission and approval.
func (m *Metrics) ObserveApprovalLead(duration time.Duration) {
	if m == nil || duration <= 0 {
		return
	}
	m.approvalLeadMs.Add(uint64(duration.Milliseconds()))
	m.approvalCount.Add(1)
	m.logger.WithField("approval_lead_ms", duration.Milliseconds()).Debug("channel approval lead observed")
}

// RecordSyncOutcome stores manual/automatic sync success ratios.
func (m *Metrics) RecordSyncOutcome(success bool) {
	if m == nil {
		return
	}
	if success {
		m.syncSuccess.Add(1)
	}
	m.syncTotal.Add(1)
}

// ObserveCredentialCoverage tracks valid vs total credentials.
func (m *Metrics) ObserveCredentialCoverage(valid bool) {
	if m == nil {
		return
	}
	if valid {
		m.credentialValid.Add(1)
	}
	m.credentialTotal.Add(1)
}

// ObserveKPILoad records histogram buckets for KPI load duration.
func (m *Metrics) ObserveKPILoad(duration time.Duration) {
	if m == nil || duration < 0 {
		return
	}
	ms := duration.Milliseconds()
	switch {
	case ms <= 1000:
		m.kpiBuckets[0].Add(1)
	case ms <= 2000:
		m.kpiBuckets[1].Add(1)
	case ms <= 3000:
		m.kpiBuckets[2].Add(1)
	default:
		m.kpiBuckets[3].Add(1)
	}
}

// Snapshot returns aggregated metrics for exporters.
func (m *Metrics) Snapshot() MetricSnapshot {
	if m == nil {
		return MetricSnapshot{}
	}
	snapshot := MetricSnapshot{}
	if count := m.approvalCount.Load(); count > 0 {
		snapshot.AverageApprovalLead = time.Duration(m.approvalLeadMs.Load()/count) * time.Millisecond
	}
	snapshot.SyncSuccessRate = rate(m.syncSuccess.Load(), m.syncTotal.Load())
	snapshot.CredentialCoverage = rate(m.credentialValid.Load(), m.credentialTotal.Load())
	snapshot.KPILoadHistogram = KPILoadHistogram{
		Under1s: m.kpiBuckets[0].Load(),
		Under2s: m.kpiBuckets[1].Load(),
		Under3s: m.kpiBuckets[2].Load(),
		Over3s:  m.kpiBuckets[3].Load(),
	}
	return snapshot
}

func rate(success, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total)
}
