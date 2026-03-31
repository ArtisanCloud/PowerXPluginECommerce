package subscription_reconciliation

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	SubscriptionReconciliationRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/subscription_reconciliation"
	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	SubscriptionReconciliationObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/subscription_reconciliation"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrServiceUnavailable  = errors.New("subscription reconciliation service unavailable")
	ErrInvalidBillingCycle = errors.New("invalid billing cycle, expected YYYY-MM-DD")
	ErrInvalidRunType      = errors.New("run type must be daily or rerun")
	ErrInvalidSLALevel     = errors.New("sla level must be high, medium or low")
	ErrInvalidResolution   = errors.New("resolution must be fixed, escalated or false_positive")
	ErrBatchNotFound       = errors.New("reconciliation batch not found")
	ErrDeltaNotFound       = errors.New("reconciliation delta not found")
	ErrTaskNotFound        = errors.New("reconciliation task not found")
	ErrTaskAlreadyOpen     = errors.New("reconciliation task already open")
	ErrNotImplemented      = errors.New("subscription reconciliation feature not implemented")
)

type Service struct {
	batchRepo     *SubscriptionReconciliationRepo.ReconciliationBatchRepository
	deltaRepo     *SubscriptionReconciliationRepo.ReconciliationDeltaRepository
	taskRepo      *SubscriptionReconciliationRepo.DeltaTaskRepository
	policyRepo    *SubscriptionReconciliationRepo.RenewalGovernancePolicyRepository
	executionRepo *SubscriptionReconciliationRepo.RenewalExecutionLogRepository
	publisher     *SubscriptionReconciliationObs.Publisher
}

type CreateBatchInput struct {
	BillingCycle string
	RunType      string
	CreatedBy    string
	Samples      []ReconciliationSample
}

type BatchListQuery struct {
	BillingCycle string
	Status       string
	Limit        int
}

type DeltaListQuery struct {
	BatchID   string
	DeltaType string
	RiskLevel string
	Limit     int
}

type CreateDeltaTaskInput struct {
	DeltaID  string
	Assignee string
	SLALevel string
	Note     string
}

type CloseTaskInput struct {
	TaskID         string
	Resolution     string
	ResolutionNote string
}

type AdjustDeltaInput struct {
	DeltaID             string
	ExpectedAmountMinor int64
	ActualAmountMinor   int64
	ReasonCode          string
	Note                string
}

type RunGovernanceInput struct {
	BillingCycle string
	DryRun       bool
}

type DashboardQuery struct {
	From string
	To   string
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{}
	}
	return &Service{
		batchRepo:     SubscriptionReconciliationRepo.NewReconciliationBatchRepository(deps.DB),
		deltaRepo:     SubscriptionReconciliationRepo.NewReconciliationDeltaRepository(deps.DB),
		taskRepo:      SubscriptionReconciliationRepo.NewDeltaTaskRepository(deps.DB),
		policyRepo:    SubscriptionReconciliationRepo.NewRenewalGovernancePolicyRepository(deps.DB),
		executionRepo: SubscriptionReconciliationRepo.NewRenewalExecutionLogRepository(deps.DB),
		publisher:     SubscriptionReconciliationObs.NewPublisher(),
	}
}

func (s *Service) ensureAvailable() error {
	if s == nil || s.batchRepo == nil || s.deltaRepo == nil || s.taskRepo == nil || s.policyRepo == nil || s.executionRepo == nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (s *Service) withTenantContext(ctx context.Context, tenantUUID string) context.Context {
	if ctx == nil {
		return nil
	}
	tid := strings.TrimSpace(tenantUUID)
	if tid == "" {
		return ctx
	}
	return AuthX.ContextWithTenantUUID(ctx, tid)
}

func normalizeRunType(v string) (string, error) {
	runType := strings.ToLower(strings.TrimSpace(v))
	if runType == "" {
		runType = "daily"
	}
	switch runType {
	case "daily", "rerun":
		return runType, nil
	default:
		return "", ErrInvalidRunType
	}
}

func normalizeSLALevel(v string) (string, error) {
	slaLevel := strings.ToLower(strings.TrimSpace(v))
	switch slaLevel {
	case "high", "medium", "low":
		return slaLevel, nil
	default:
		return "", ErrInvalidSLALevel
	}
}

func normalizeResolution(v string) (string, error) {
	resolution := strings.ToLower(strings.TrimSpace(v))
	switch resolution {
	case "fixed", "escalated", "false_positive":
		return resolution, nil
	default:
		return "", ErrInvalidResolution
	}
}

func validateBillingCycle(v string) (string, error) {
	cycle := strings.TrimSpace(v)
	if cycle == "" {
		return "", ErrInvalidBillingCycle
	}
	if _, err := time.Parse("2006-01-02", cycle); err != nil {
		return "", ErrInvalidBillingCycle
	}
	return cycle, nil
}

func fingerprintForSample(cycle string, sample ReconciliationSample, deltaType string) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%d|%d", cycle, sample.SubscriptionRef, sample.BillRef, deltaType, sample.ExpectedAmountMinor, sample.ActualAmountMinor)
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func slaDeadline(now time.Time, level string) time.Time {
	switch level {
	case "high":
		return now.Add(24 * time.Hour)
	case "medium":
		return now.Add(48 * time.Hour)
	default:
		return now.Add(72 * time.Hour)
	}
}

func (s *Service) CreateBatch(ctx context.Context, tenantUUID string, input CreateBatchInput) (*SubscriptionReconciliationModel.ReconciliationBatch, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	cycle, err := validateBillingCycle(input.BillingCycle)
	if err != nil {
		return nil, err
	}
	runType, err := normalizeRunType(input.RunType)
	if err != nil {
		return nil, err
	}
	ctx = s.withTenantContext(ctx, tenantUUID)

	existing, err := s.batchRepo.FindByCycleRunType(ctx, cycle, runType)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now().UTC()
	batch := &SubscriptionReconciliationModel.ReconciliationBatch{
		ID:           utils.NewUUID(),
		BillingCycle: cycle,
		RunType:      runType,
		Status:       "completed",
		StartedAt:    &now,
		FinishedAt:   &now,
		CreatedBy:    strings.TrimSpace(input.CreatedBy),
	}

	deltas := make([]SubscriptionReconciliationModel.ReconciliationDelta, 0, len(input.Samples))
	for _, sample := range input.Samples {
		batch.ExpectedAmountMinor += sample.ExpectedAmountMinor
		batch.ActualAmountMinor += sample.ActualAmountMinor
		deltaType, riskLevel, deltaAmount := classifyDelta(sample)
		if deltaType == "" {
			continue
		}
		batch.DeltaCount++
		batch.DeltaAmountMinor += deltaAmount
		deltas = append(deltas, SubscriptionReconciliationModel.ReconciliationDelta{
			ID:                  utils.NewUUID(),
			BatchID:             batch.ID,
			SubscriptionRef:     strings.TrimSpace(sample.SubscriptionRef),
			BillRef:             strings.TrimSpace(sample.BillRef),
			PaymentRef:          strings.TrimSpace(sample.PaymentRef),
			DeltaType:           deltaType,
			RiskLevel:           riskLevel,
			ExpectedAmountMinor: sample.ExpectedAmountMinor,
			ActualAmountMinor:   sample.ActualAmountMinor,
			DeltaAmountMinor:    deltaAmount,
			ReasonCode:          strings.TrimSpace(sample.ReasonCode),
			Status:              "open",
			DeltaFingerprint:    fingerprintForSample(cycle, sample, deltaType),
			DetectedAt:          &now,
		})
	}
	batch.DeltaAmountMinor = batch.ExpectedAmountMinor - batch.ActualAmountMinor

	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, err
	}
	if err := s.deltaRepo.CreateBatch(ctx, deltas); err != nil {
		return nil, err
	}

	s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventReconciliationGenerated, map[string]any{
		"batch_id":      batch.ID,
		"billing_cycle": batch.BillingCycle,
		"delta_count":   batch.DeltaCount,
	})
	return batch, nil
}

func (s *Service) ListBatches(ctx context.Context, tenantUUID string, query BatchListQuery) ([]SubscriptionReconciliationModel.ReconciliationBatch, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	ctx = s.withTenantContext(ctx, tenantUUID)
	return s.batchRepo.List(ctx, query.BillingCycle, query.Status, query.Limit)
}

func (s *Service) ListDeltas(ctx context.Context, tenantUUID string, query DeltaListQuery) ([]SubscriptionReconciliationModel.ReconciliationDelta, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(query.BatchID) == "" {
		return nil, ErrBatchNotFound
	}
	ctx = s.withTenantContext(ctx, tenantUUID)
	return s.deltaRepo.ListByBatchID(ctx, query.BatchID, query.DeltaType, query.RiskLevel, query.Limit)
}

func (s *Service) CreateDeltaTask(ctx context.Context, tenantUUID string, input CreateDeltaTaskInput) (*SubscriptionReconciliationModel.DeltaTask, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	slaLevel, err := normalizeSLALevel(input.SLALevel)
	if err != nil {
		return nil, err
	}
	ctx = s.withTenantContext(ctx, tenantUUID)

	delta, err := s.deltaRepo.GetByID(ctx, input.DeltaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeltaNotFound
		}
		return nil, err
	}
	openTask, err := s.taskRepo.FindOpenByFingerprint(ctx, delta.DeltaFingerprint)
	if err == nil && openTask != nil {
		return nil, ErrTaskAlreadyOpen
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	deadline := slaDeadline(now, slaLevel)
	task := &SubscriptionReconciliationModel.DeltaTask{
		ID:               utils.NewUUID(),
		DeltaID:          delta.ID,
		DeltaFingerprint: delta.DeltaFingerprint,
		Assignee:         strings.TrimSpace(input.Assignee),
		Priority:         slaLevel,
		SLALevel:         slaLevel,
		SLADeadline:      &deadline,
		Status:           "pending",
		ResolutionNote:   strings.TrimSpace(input.Note),
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventDeltaTaskCreated, map[string]any{
		"task_id":  task.ID,
		"delta_id": delta.ID,
	})
	return task, nil
}

func (s *Service) CloseTask(ctx context.Context, tenantUUID string, input CloseTaskInput) (*SubscriptionReconciliationModel.DeltaTask, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	resolution, err := normalizeResolution(input.Resolution)
	if err != nil {
		return nil, err
	}
	ctx = s.withTenantContext(ctx, tenantUUID)
	now := time.Now().UTC()
	task, err := s.taskRepo.Close(ctx, input.TaskID, resolution, input.ResolutionNote, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventDeltaTaskClosed, map[string]any{
		"task_id":    task.ID,
		"resolution": task.Resolution,
		"closed_at":  now.Format(time.RFC3339Nano),
	})
	return task, nil
}

func (s *Service) AdjustDelta(ctx context.Context, tenantUUID string, input AdjustDeltaInput) (*SubscriptionReconciliationModel.ReconciliationDelta, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	ctx = s.withTenantContext(ctx, tenantUUID)
	delta, err := s.deltaRepo.GetByID(ctx, input.DeltaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeltaNotFound
		}
		return nil, err
	}
	delta.ExpectedAmountMinor = input.ExpectedAmountMinor
	delta.ActualAmountMinor = input.ActualAmountMinor
	delta.DeltaAmountMinor = input.ExpectedAmountMinor - input.ActualAmountMinor
	if strings.TrimSpace(input.ReasonCode) != "" {
		delta.ReasonCode = strings.TrimSpace(input.ReasonCode)
	}
	delta.Status = "processing"
	if err := s.deltaRepo.Save(ctx, delta); err != nil {
		return nil, err
	}
	s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventDeltaAdjusted, map[string]any{
		"delta_id":    delta.ID,
		"reason_code": delta.ReasonCode,
		"note":        strings.TrimSpace(input.Note),
	})
	return delta, nil
}

func (s *Service) RunGovernance(ctx context.Context, tenantUUID string, input RunGovernanceInput) (map[string]any, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	_ = s.withTenantContext(ctx, tenantUUID)
	_ = input
	return nil, ErrNotImplemented
}

func (s *Service) Dashboard(ctx context.Context, tenantUUID string, query DashboardQuery) (map[string]any, error) {
	if err := s.ensureAvailable(); err != nil {
		return nil, err
	}
	_ = s.withTenantContext(ctx, tenantUUID)
	_ = query
	return nil, ErrNotImplemented
}
