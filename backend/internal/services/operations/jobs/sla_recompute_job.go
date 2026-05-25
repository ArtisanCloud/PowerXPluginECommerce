package jobs

import (
	"context"
	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"time"

	operationsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/operations"
	"github.com/sirupsen/logrus"
)

// SLARecomputeJob triggers SLA score recomputation across profiles.
type SLARecomputeJob struct {
	service *operationsvc.SLAService
	log     *logrus.Entry
}

// NewSLARecomputeJob constructs a job instance.
func NewSLARecomputeJob(service *operationsvc.SLAService, log *logrus.Entry) *SLARecomputeJob {
	if log == nil {
		log = pxlogger.WithField("component", "operations.sla_recompute_job")
	}
	return &SLARecomputeJob{service: service, log: log}
}

// Run executes the recomputation workflow.
func (j *SLARecomputeJob) Run(ctx context.Context) error {
	if j.service == nil {
		return nil
	}
	start := time.Now()
	profiles, err := j.service.RecomputeScores(ctx)
	if err != nil {
		return err
	}
	j.log.WithContext(ctx).WithFields(logrus.Fields{
		"profiles": len(profiles),
		"duration": time.Since(start).String(),
	}).Info("recomputed SLA profiles")
	return nil
}
