package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

type GatewayRecoveryService struct {
	failureRepo *LogisticsRepo.GatewayFailureEventRepository
	jobRepo     *LogisticsRepo.TrackingSyncJobRepository
	waybillRepo *LogisticsRepo.WaybillRepository
	waybillSvc  *WaybillService
}

type GatewayFailureQuery struct {
	CarrierID string `json:"carrier_id,omitempty"`
	WaybillNo string `json:"waybill_no,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func NewGatewayRecoveryService(deps *app.Deps) *GatewayRecoveryService {
	if deps == nil || deps.DB == nil {
		return &GatewayRecoveryService{}
	}
	return &GatewayRecoveryService{
		failureRepo: LogisticsRepo.NewGatewayFailureEventRepository(deps.DB),
		jobRepo:     LogisticsRepo.NewTrackingSyncJobRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		waybillSvc:  NewWaybillService(deps),
	}
}

func (s *GatewayRecoveryService) IngestFromFailedJobs(ctx context.Context, tenantUUID string, hours int) (int, error) {
	if s == nil || s.failureRepo == nil || s.jobRepo == nil {
		return 0, errors.New("gateway recovery service unavailable")
	}
	if hours <= 0 {
		hours = 24
	}
	ctx = withTenantContext(ctx, tenantUUID)
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	jobs, err := s.jobRepo.List(ctx, LogisticsRepo.TrackingSyncJobFilter{
		Status: "failed",
		Limit:  200,
		Since:  &since,
	})
	if err != nil {
		return 0, err
	}
	partial, err := s.jobRepo.List(ctx, LogisticsRepo.TrackingSyncJobFilter{
		Status: "partial_failed",
		Limit:  200,
		Since:  &since,
	})
	if err == nil {
		jobs = append(jobs, partial...)
	}
	created := 0
	for _, row := range jobs {
		if strings.TrimSpace(row.LastError) == "" {
			continue
		}
		class, code := classifyGatewayError(row.LastError)
		event := &LogisticsModel.GatewayFailureEvent{
			ID:           utils.NewUUID(),
			CarrierID:    strings.TrimSpace(row.CarrierID),
			Provider:     strings.TrimSpace(row.Provider),
			SourceJobID:  row.ID,
			ErrorClass:   class,
			ErrorCode:    code,
			ErrorMessage: strings.TrimSpace(row.LastError),
			Status:       "pending",
			RetryCount:   0,
			Metadata:     []byte("{}"),
		}
		if class == "auth" {
			openTill := time.Now().UTC().Add(5 * time.Minute)
			event.CircuitOpenTill = &openTill
		}
		if err := s.failureRepo.Create(ctx, event); err == nil {
			created++
		}
	}
	return created, nil
}

func (s *GatewayRecoveryService) List(ctx context.Context, tenantUUID string, query GatewayFailureQuery) ([]LogisticsModel.GatewayFailureEvent, error) {
	if s == nil || s.failureRepo == nil {
		return nil, errors.New("gateway recovery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	_, _ = s.IngestFromFailedJobs(ctx, tenantUUID, 24)
	return s.failureRepo.List(ctx, LogisticsRepo.GatewayFailureEventFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		WaybillNo: strings.TrimSpace(query.WaybillNo),
		Status:    strings.TrimSpace(query.Status),
		Limit:     query.Limit,
	})
}

func (s *GatewayRecoveryService) Compensate(ctx context.Context, tenantUUID, id string) (*LogisticsModel.GatewayFailureEvent, error) {
	if s == nil || s.failureRepo == nil || s.waybillRepo == nil || s.waybillSvc == nil {
		return nil, errors.New("gateway recovery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.failureRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if row.CircuitOpenTill != nil && now.Before(*row.CircuitOpenTill) {
		return nil, errors.New("circuit breaker open")
	}
	if strings.TrimSpace(row.WaybillID) == "" && strings.TrimSpace(row.WaybillNo) != "" {
		if wb, wbErr := s.waybillRepo.GetByWaybillNo(ctx, row.WaybillNo); wbErr == nil && wb != nil {
			row.WaybillID = wb.ID
		}
	}
	if strings.TrimSpace(row.WaybillID) == "" {
		return nil, errors.New("waybill not found for compensation")
	}
	_, syncErr := s.waybillSvc.SyncTrackingFromProvider(ctx, tenantUUID, row.WaybillID, 20)
	row.RetryCount++
	if syncErr != nil {
		row.ErrorMessage = syncErr.Error()
		class, code := classifyGatewayError(row.ErrorMessage)
		row.ErrorClass = class
		row.ErrorCode = code
		row.Status = "pending"
		next := now.Add(time.Duration(row.RetryCount*30) * time.Second)
		row.NextRetryAt = &next
		if class == "auth" {
			openTill := now.Add(5 * time.Minute)
			row.CircuitOpenTill = &openTill
		}
		if err := s.failureRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, syncErr
	}
	row.Status = "recovered"
	row.ErrorMessage = ""
	row.NextRetryAt = nil
	row.CircuitOpenTill = nil
	recoveredAt := now
	row.RecoveredAt = &recoveredAt
	if err := s.failureRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func classifyGatewayError(message string) (string, string) {
	msg := strings.ToLower(strings.TrimSpace(message))
	switch {
	case strings.Contains(msg, "401"), strings.Contains(msg, "authorization"), strings.Contains(msg, "token"):
		return "auth", "E_AUTH"
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return "timeout", "E_TIMEOUT"
	case strings.Contains(msg, "status=5"), strings.Contains(msg, "bad gateway"):
		return "server_5xx", "E_5XX"
	case strings.Contains(msg, "status=4"), strings.Contains(msg, "bad request"):
		return "client_4xx", "E_4XX"
	case strings.Contains(msg, "contract"), strings.Contains(msg, "decode"):
		return "contract", "E_CONTRACT"
	default:
		return "unknown", "E_UNKNOWN"
	}
}
