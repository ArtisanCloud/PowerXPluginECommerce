package payments

import (
	"context"
	"errors"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

var (
	ErrReconciliationServiceUnavailable = errors.New("reconciliation service unavailable")
	ErrReconciliationInvalidPeriod      = errors.New("invalid reconciliation period")
	ErrReconciliationItemsRequired      = errors.New("reconciliation items required")
	ErrReconciliationNotFound           = errors.New("reconciliation not found")
	ErrReconciliationItemNotFound       = errors.New("reconciliation item not found")
)

type ReconciliationService struct {
	deps          *app.Deps
	repo          *paymentrepo.PaymentReconciliationRepository
	itemRepo      *paymentrepo.PaymentReconciliationItemRepository
	logger        *paymentslogger.Logger
}

func NewReconciliationService(deps *app.Deps) *ReconciliationService {
	if deps == nil || deps.DB == nil {
		return &ReconciliationService{deps: deps}
	}
	var obsLogger *paymentslogger.Logger
	if entry := deps.RuntimeLogger(deps.Ctx, "payments", nil); entry != nil {
		obsLogger = paymentslogger.NewLogger(entry)
	}
	return &ReconciliationService{
		deps:     deps,
		repo:     paymentrepo.NewPaymentReconciliationRepository(deps.DB),
		itemRepo: paymentrepo.NewPaymentReconciliationItemRepository(deps.DB),
		logger:   obsLogger,
	}
}

func (s *ReconciliationService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil && s.itemRepo != nil
}

func (s *ReconciliationService) ListReconciliations(ctx context.Context, tenantUUID, adminID string) ([]ReconciliationDTO, error) {
	if !s.Ready() {
		return nil, ErrReconciliationServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var rows []models.PaymentReconciliation
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("period_start DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ReconciliationDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toReconciliationDTO(&rows[i]))
	}
	return out, nil
}

func (s *ReconciliationService) CreateReconciliation(
	ctx context.Context,
	tenantUUID, adminID string,
	req CreateReconciliationRequest,
) (*ReconciliationDTO, error) {
	if !s.Ready() {
		return nil, ErrReconciliationServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if strings.TrimSpace(req.PeriodType) == "" || req.PeriodStart.IsZero() || req.PeriodEnd.IsZero() {
		return nil, ErrReconciliationInvalidPeriod
	}
	if req.PeriodStart.After(req.PeriodEnd) {
		return nil, ErrReconciliationInvalidPeriod
	}
	if len(req.Items) == 0 {
		return nil, ErrReconciliationItemsRequired
	}

	diffCount := len(req.Items)
	var diffTotal int64
	for _, item := range req.Items {
		diffTotal += item.DiffAmount
	}

	tx := s.deps.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() { _ = tx.Rollback() }()

	reconciliation := models.PaymentReconciliation{
		BaseModel: models.BaseModel{TenantUuid: tenantUUID},
		PeriodType: req.PeriodType,
		PeriodStart: req.PeriodStart.UTC(),
		PeriodEnd: req.PeriodEnd.UTC(),
		DiffCount: diffCount,
		DiffTotalAmount: diffTotal,
		Status: "pending",
	}
	if err := tx.Create(&reconciliation).Error; err != nil {
		return nil, err
	}

	items := make([]models.PaymentReconciliationItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, models.PaymentReconciliationItem{
			BaseModel: models.BaseModel{TenantUuid: tenantUUID},
			ReconciliationID: reconciliation.ID,
			TransactionID: item.TransactionID,
			DiffType: strings.TrimSpace(item.DiffType),
			DiffAmount: item.DiffAmount,
		})
	}
	if err := tx.Create(&items).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	if s.logger != nil {
		requestID, _ := authx.RequestIDFromContext(ctx)
		s.logger.EmitEvent(paymentslogger.Event{
			Action:    "reconciliation.created",
			TenantID:  tenantUUID,
			ActorID:   adminID,
			RequestID: requestID,
			Status:    reconciliation.Status,
			Metadata: map[string]any{
				"reconciliation_id": reconciliation.ID,
				"period_type":       reconciliation.PeriodType,
				"period_start":      reconciliation.PeriodStart,
				"period_end":        reconciliation.PeriodEnd,
				"diff_count":        reconciliation.DiffCount,
				"diff_total_amount": reconciliation.DiffTotalAmount,
			},
		})
	}
	out := toReconciliationDTO(&reconciliation)
	return &out, nil
}

func (s *ReconciliationService) ListReconciliationItems(
	ctx context.Context,
	tenantUUID, adminID string,
	reconciliationID uint64,
) ([]ReconciliationItemDTO, error) {
	if !s.Ready() {
		return nil, ErrReconciliationServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var rows []models.PaymentReconciliationItem
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND reconciliation_id = ?", tenantUUID, reconciliationID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ReconciliationItemDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toReconciliationItemDTO(&rows[i]))
	}
	return out, nil
}

func (s *ReconciliationService) ResolveReconciliationItem(
	ctx context.Context,
	tenantUUID, adminID string,
	reconciliationID, itemID uint64,
	req ResolveReconciliationItemRequest,
) (*ReconciliationItemDTO, error) {
	if !s.Ready() {
		return nil, ErrReconciliationServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	now := time.Now().UTC()

	tx := s.deps.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() { _ = tx.Rollback() }()

	var item models.PaymentReconciliationItem
	if err := tx.Where("tenant_uuid = ? AND reconciliation_id = ? AND id = ?", tenantUUID, reconciliationID, itemID).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReconciliationItemNotFound
		}
		return nil, err
	}
	item.Resolution = strings.TrimSpace(req.Resolution)
	item.ResolvedBy = strings.TrimSpace(adminID)
	item.ResolvedAt = &now
	if err := tx.Save(&item).Error; err != nil {
		return nil, err
	}

	var remaining int64
	if err := tx.Model(&models.PaymentReconciliationItem{}).
		Where("tenant_uuid = ? AND reconciliation_id = ? AND resolved_at IS NULL", tenantUUID, reconciliationID).
		Count(&remaining).Error; err != nil {
		return nil, err
	}
	if remaining == 0 {
		if err := tx.Model(&models.PaymentReconciliation{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, reconciliationID).
			Updates(map[string]interface{}{
				"status":       "completed",
				"processed_by": adminID,
				"processed_at": now,
			}).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	if s.logger != nil {
		requestID, _ := authx.RequestIDFromContext(ctx)
		s.logger.EmitEvent(paymentslogger.Event{
			Action:    "reconciliation.item.resolved",
			TenantID:  tenantUUID,
			ActorID:   adminID,
			RequestID: requestID,
			Status:    "resolved",
			Metadata: map[string]any{
				"reconciliation_id": reconciliationID,
				"item_id":           item.ID,
				"diff_type":         item.DiffType,
				"diff_amount":       item.DiffAmount,
				"resolution":        item.Resolution,
			},
		})
	}
	out := toReconciliationItemDTO(&item)
	return &out, nil
}

func toReconciliationDTO(row *models.PaymentReconciliation) ReconciliationDTO {
	if row == nil {
		return ReconciliationDTO{}
	}
	return ReconciliationDTO{
		ID:              row.ID,
		PeriodType:      row.PeriodType,
		PeriodStart:     row.PeriodStart,
		PeriodEnd:       row.PeriodEnd,
		DiffCount:       row.DiffCount,
		DiffTotalAmount: row.DiffTotalAmount,
		Status:          row.Status,
		ProcessedBy:     row.ProcessedBy,
		ProcessedAt:     row.ProcessedAt,
		CreatedAt:       row.CreatedAt,
	}
}

func toReconciliationItemDTO(row *models.PaymentReconciliationItem) ReconciliationItemDTO {
	if row == nil {
		return ReconciliationItemDTO{}
	}
	return ReconciliationItemDTO{
		ID:               row.ID,
		ReconciliationID: row.ReconciliationID,
		TransactionID:    row.TransactionID,
		DiffType:         row.DiffType,
		DiffAmount:       row.DiffAmount,
		Resolution:       row.Resolution,
		ResolvedBy:       row.ResolvedBy,
		ResolvedAt:       row.ResolvedAt,
		CreatedAt:        row.CreatedAt,
	}
}
