package logistics

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

type TrackingSyncJobService struct {
	jobRepo     *LogisticsRepo.TrackingSyncJobRepository
	waybillRepo *LogisticsRepo.WaybillRepository
	waybillSvc  *WaybillService
}

type CreateTrackingSyncJobRequest struct {
	CarrierID     string `json:"carrier_id,omitempty"`
	WaybillStatus string `json:"waybill_status,omitempty"`
	BatchLimit    int    `json:"batch_limit,omitempty"`
	EventLimit    int    `json:"event_limit,omitempty"`
}

type TrackingSyncJobQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func NewTrackingSyncJobService(deps *app.Deps) *TrackingSyncJobService {
	if deps == nil || deps.DB == nil {
		return &TrackingSyncJobService{}
	}
	return &TrackingSyncJobService{
		jobRepo:     LogisticsRepo.NewTrackingSyncJobRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		waybillSvc:  NewWaybillService(deps),
	}
}

func (s *TrackingSyncJobService) CreateAndRun(ctx context.Context, tenantUUID string, req CreateTrackingSyncJobRequest) (*LogisticsModel.TrackingSyncJob, error) {
	if s == nil || s.jobRepo == nil || s.waybillRepo == nil || s.waybillSvc == nil {
		return nil, errors.New("tracking sync job service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	batchLimit := req.BatchLimit
	if batchLimit <= 0 {
		batchLimit = 20
	}
	if batchLimit > 500 {
		batchLimit = 500
	}
	eventLimit := req.EventLimit
	if eventLimit <= 0 {
		eventLimit = 20
	}
	if eventLimit > 200 {
		eventLimit = 200
	}
	job := &LogisticsModel.TrackingSyncJob{
		ID:            utils.NewUUID(),
		CarrierID:     strings.TrimSpace(req.CarrierID),
		WaybillStatus: strings.TrimSpace(strings.ToLower(req.WaybillStatus)),
		Status:        "pending",
		BatchLimit:    batchLimit,
		EventLimit:    eventLimit,
		Metadata:      []byte("{}"),
	}
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}
	started := time.Now().UTC()
	job.StartedAt = &started
	job.Status = "running"
	_ = s.jobRepo.Save(ctx, job)

	waybills, err := s.waybillRepo.ListForSync(ctx, LogisticsRepo.WaybillSyncFilter{
		CarrierID: job.CarrierID,
		Status:    job.WaybillStatus,
		Limit:     job.BatchLimit,
	})
	if err != nil {
		job.Status = "failed"
		job.LastError = err.Error()
		finished := time.Now().UTC()
		job.FinishedAt = &finished
		_ = s.jobRepo.Save(ctx, job)
		return nil, err
	}
	job.TotalWaybills = len(waybills)
	latencies := make([]int, 0, len(waybills))
	for _, wb := range waybills {
		callStart := time.Now()
		result, syncErr := s.waybillSvc.SyncTrackingFromProvider(ctx, tenantUUID, wb.ID, job.EventLimit)
		latency := int(time.Since(callStart).Milliseconds())
		if latency < 0 {
			latency = 0
		}
		latencies = append(latencies, latency)
		if syncErr != nil {
			job.FailedCount++
			job.LastError = syncErr.Error()
			continue
		}
		job.SuccessCount++
		if result != nil {
			job.AppendedCount += result.Appended
			job.ReplayedCount += result.Replayed
		}
	}
	job.P95LatencyMS = percentile95(latencies)
	if job.FailedCount == 0 {
		job.Status = "success"
	} else if job.SuccessCount == 0 {
		job.Status = "failed"
	} else {
		job.Status = "partial_failed"
	}
	finished := time.Now().UTC()
	job.FinishedAt = &finished
	if err := s.jobRepo.Save(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *TrackingSyncJobService) List(ctx context.Context, tenantUUID string, query TrackingSyncJobQuery) ([]LogisticsModel.TrackingSyncJob, error) {
	if s == nil || s.jobRepo == nil {
		return nil, errors.New("tracking sync job service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	return s.jobRepo.List(ctx, LogisticsRepo.TrackingSyncJobFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		Status:    strings.TrimSpace(strings.ToLower(query.Status)),
		Limit:     query.Limit,
	})
}

func (s *TrackingSyncJobService) Cancel(ctx context.Context, tenantUUID, id, reason string) (*LogisticsModel.TrackingSyncJob, error) {
	if s == nil || s.jobRepo == nil {
		return nil, errors.New("tracking sync job service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == "success" || row.Status == "failed" || row.Status == "partial_failed" {
		return row, nil
	}
	row.Status = "cancelled"
	row.CancelledReason = strings.TrimSpace(reason)
	finished := time.Now().UTC()
	row.FinishedAt = &finished
	if err := s.jobRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *TrackingSyncJobService) Retry(ctx context.Context, tenantUUID, id string) (*LogisticsModel.TrackingSyncJob, error) {
	if s == nil || s.jobRepo == nil {
		return nil, errors.New("tracking sync job service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.CreateAndRun(ctx, tenantUUID, CreateTrackingSyncJobRequest{
		CarrierID:     row.CarrierID,
		WaybillStatus: row.WaybillStatus,
		BatchLimit:    row.BatchLimit,
		EventLimit:    row.EventLimit,
	})
}

func percentile95(values []int) int {
	if len(values) == 0 {
		return 0
	}
	clean := make([]int, 0, len(values))
	for _, v := range values {
		if v >= 0 {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return 0
	}
	sort.Ints(clean)
	idx := int(float64(len(clean)-1) * 0.95)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(clean) {
		idx = len(clean) - 1
	}
	return clean[idx]
}
