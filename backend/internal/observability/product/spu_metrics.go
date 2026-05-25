package product

import (
	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// SPUMetrics tracks KPI style counters for SPU flows.
type SPUMetrics struct {
	logger *logrus.Entry

	leadTimeMs atomic.Uint64
	leadCount  atomic.Uint64

	importSuccess atomic.Uint64
	importTotal   atomic.Uint64

	channelSuccess atomic.Uint64
	channelTotal   atomic.Uint64

	approvalOnTime atomic.Uint64
	approvalTotal  atomic.Uint64
}

// SPUMetricSnapshot exposes aggregated values for dashboards.
type SPUMetricSnapshot struct {
	AverageLeadTime    time.Duration `json:"averageLeadTime"`
	ImportSuccessRate  float64       `json:"importSuccessRate"`
	ChannelSuccessRate float64       `json:"channelSuccessRate"`
	ApprovalOnTimeRate float64       `json:"approvalOnTimeRate"`
}

// NewSPUMetrics creates a collector with optional logger.
func NewSPUMetrics(logger *logrus.Entry) *SPUMetrics {
	if logger == nil {
		logger = pxlogger.WithField("component", "spu-metrics")
	}
	return &SPUMetrics{logger: logger}
}

// ObserveLeadTime stores durations from draft to publish.
func (m *SPUMetrics) ObserveLeadTime(duration time.Duration) {
	if m == nil || duration <= 0 {
		return
	}
	m.leadTimeMs.Add(uint64(duration.Milliseconds()))
	m.leadCount.Add(1)
	if m.logger != nil {
		m.logger.WithField("lead_time_ms", duration.Milliseconds()).Debug("spu lead time observed")
	}
}

// RecordImport updates import success/failure counters.
func (m *SPUMetrics) RecordImport(successCount, failureCount int) {
	if m == nil {
		return
	}
	total := successCount + failureCount
	if total <= 0 {
		return
	}
	m.importSuccess.Add(uint64(successCount))
	m.importTotal.Add(uint64(total))
	if m.logger != nil {
		m.logger.WithFields(logrus.Fields{
			"success": successCount,
			"failed":  failureCount,
		}).Debug("spu import batch recorded")
	}
}

// RecordChannelSync stores per-channel publish outcomes.
func (m *SPUMetrics) RecordChannelSync(success bool) {
	if m == nil {
		return
	}
	if success {
		m.channelSuccess.Add(1)
	}
	m.channelTotal.Add(1)
}

// RecordApprovalSLA notes whether an approval finished before SLA.
func (m *SPUMetrics) RecordApprovalSLA(onTime bool) {
	if m == nil {
		return
	}
	if onTime {
		m.approvalOnTime.Add(1)
	}
	m.approvalTotal.Add(1)
}

// Snapshot returns averages and rates for dashboards.
func (m *SPUMetrics) Snapshot() SPUMetricSnapshot {
	if m == nil {
		return SPUMetricSnapshot{}
	}
	snapshot := SPUMetricSnapshot{}
	if count := m.leadCount.Load(); count > 0 {
		ms := m.leadTimeMs.Load()
		snapshot.AverageLeadTime = time.Duration(ms/count) * time.Millisecond
	}
	snapshot.ImportSuccessRate = calculateRate(m.importSuccess.Load(), m.importTotal.Load())
	snapshot.ChannelSuccessRate = calculateRate(m.channelSuccess.Load(), m.channelTotal.Load())
	snapshot.ApprovalOnTimeRate = calculateRate(m.approvalOnTime.Load(), m.approvalTotal.Load())
	return snapshot
}

func calculateRate(success, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total)
}
