package logistics

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

type TrackingSyncSchedulerService struct {
	scheduleRepo *LogisticsRepo.TrackingSyncScheduleRepository
	jobSvc       *TrackingSyncJobService
}

type UpsertTrackingSyncScheduleRequest struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	CronExpr        string `json:"cron_expr"`
	CarrierID       string `json:"carrier_id,omitempty"`
	WaybillStatus   string `json:"waybill_status,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
	MaxConcurrency  int    `json:"max_concurrency,omitempty"`
	DedupeWindowSec int    `json:"dedupe_window_sec,omitempty"`
	BatchLimit      int    `json:"batch_limit,omitempty"`
	EventLimit      int    `json:"event_limit,omitempty"`
}

func NewTrackingSyncSchedulerService(deps *app.Deps) *TrackingSyncSchedulerService {
	if deps == nil || deps.DB == nil {
		return &TrackingSyncSchedulerService{}
	}
	return &TrackingSyncSchedulerService{
		scheduleRepo: LogisticsRepo.NewTrackingSyncScheduleRepository(deps.DB),
		jobSvc:       NewTrackingSyncJobService(deps),
	}
}

func (s *TrackingSyncSchedulerService) Upsert(ctx context.Context, tenantUUID string, req UpsertTrackingSyncScheduleRequest) (*LogisticsModel.TrackingSyncSchedule, error) {
	if s == nil || s.scheduleRepo == nil {
		return nil, errors.New("tracking sync scheduler service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	cronExpr := strings.TrimSpace(req.CronExpr)
	if cronExpr == "" {
		return nil, errors.New("cron_expr is required")
	}
	now := time.Now().UTC()
	var row *LogisticsModel.TrackingSyncSchedule
	if strings.TrimSpace(req.ID) != "" {
		existing, err := s.scheduleRepo.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		row = existing
	} else {
		row = &LogisticsModel.TrackingSyncSchedule{
			ID:       utils.NewUUID(),
			Metadata: []byte("{}"),
		}
	}
	row.Name = name
	row.CronExpr = cronExpr
	row.CarrierID = strings.TrimSpace(req.CarrierID)
	row.WaybillStatus = strings.TrimSpace(strings.ToLower(req.WaybillStatus))
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	} else if strings.TrimSpace(req.ID) == "" {
		row.Enabled = true
	}
	if req.MaxConcurrency > 0 {
		row.MaxConcurrency = req.MaxConcurrency
	} else if row.MaxConcurrency <= 0 {
		row.MaxConcurrency = 1
	}
	if req.DedupeWindowSec > 0 {
		row.DedupeWindowSec = req.DedupeWindowSec
	} else if row.DedupeWindowSec <= 0 {
		row.DedupeWindowSec = 300
	}
	if req.BatchLimit > 0 {
		row.BatchLimit = req.BatchLimit
	} else if row.BatchLimit <= 0 {
		row.BatchLimit = 20
	}
	if req.EventLimit > 0 {
		row.EventLimit = req.EventLimit
	} else if row.EventLimit <= 0 {
		row.EventLimit = 20
	}
	if row.NextTriggerAt == nil {
		next := nextRunAt(now, cronExpr)
		row.NextTriggerAt = &next
	}

	if strings.TrimSpace(req.ID) == "" {
		if err := s.scheduleRepo.Create(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	if err := s.scheduleRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *TrackingSyncSchedulerService) List(ctx context.Context, tenantUUID string, enabled *bool, limit int) ([]LogisticsModel.TrackingSyncSchedule, error) {
	if s == nil || s.scheduleRepo == nil {
		return nil, errors.New("tracking sync scheduler service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	return s.scheduleRepo.List(ctx, enabled, limit)
}

func (s *TrackingSyncSchedulerService) Toggle(ctx context.Context, tenantUUID, id string, enabled bool) (*LogisticsModel.TrackingSyncSchedule, error) {
	if s == nil || s.scheduleRepo == nil {
		return nil, errors.New("tracking sync scheduler service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	row.Enabled = enabled
	now := time.Now().UTC()
	if enabled {
		next := nextRunAt(now, row.CronExpr)
		row.NextTriggerAt = &next
	}
	if err := s.scheduleRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *TrackingSyncSchedulerService) TriggerNow(ctx context.Context, tenantUUID, id string) (*LogisticsModel.TrackingSyncJob, error) {
	if s == nil || s.scheduleRepo == nil || s.jobSvc == nil {
		return nil, errors.New("tracking sync scheduler service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.jobSvc.CreateAndRun(ctx, tenantUUID, CreateTrackingSyncJobRequest{
		CarrierID:     row.CarrierID,
		WaybillStatus: row.WaybillStatus,
		BatchLimit:    row.BatchLimit,
		EventLimit:    row.EventLimit,
	})
}

func (s *TrackingSyncSchedulerService) RunDue(ctx context.Context, tenantUUID string, limit int) (int, error) {
	if s == nil || s.scheduleRepo == nil || s.jobSvc == nil {
		return 0, errors.New("tracking sync scheduler service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	now := time.Now().UTC()
	rows, err := s.scheduleRepo.ListDue(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	triggered := 0
	for i := range rows {
		row := &rows[i]
		if !row.Enabled {
			continue
		}
		if row.LastTriggeredAt != nil && row.DedupeWindowSec > 0 {
			if now.Sub(*row.LastTriggeredAt) < time.Duration(row.DedupeWindowSec)*time.Second {
				continue
			}
		}
		_, runErr := s.jobSvc.CreateAndRun(ctx, tenantUUID, CreateTrackingSyncJobRequest{
			CarrierID:     row.CarrierID,
			WaybillStatus: row.WaybillStatus,
			BatchLimit:    row.BatchLimit,
			EventLimit:    row.EventLimit,
		})
		last := now
		row.LastTriggeredAt = &last
		next := nextRunAt(now, row.CronExpr)
		row.NextTriggerAt = &next
		if runErr != nil {
			_ = s.scheduleRepo.Save(ctx, row)
			continue
		}
		if err := s.scheduleRepo.Save(ctx, row); err == nil {
			triggered++
		}
	}
	return triggered, nil
}

func nextRunAt(base time.Time, cronExpr string) time.Time {
	interval := cronIntervalMinutes(cronExpr)
	return base.Add(time.Duration(interval) * time.Minute).UTC()
}

func cronIntervalMinutes(expr string) int {
	v := strings.TrimSpace(expr)
	if strings.HasPrefix(v, "*/") {
		token := strings.TrimPrefix(v, "*/")
		if idx := strings.Index(token, " "); idx > 0 {
			token = token[:idx]
		}
		if m, err := strconv.Atoi(token); err == nil && m > 0 {
			return m
		}
	}
	return 15
}
