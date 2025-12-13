package product

import "github.com/sirupsen/logrus"

// Metrics aggregates observability helpers for the product domain.
type Metrics struct {
	SPU *SPUMetrics
}

// NewMetrics wires product metrics with a shared logger.
func NewMetrics(logger *logrus.Entry) *Metrics {
	return &Metrics{SPU: NewSPUMetrics(logger)}
}
