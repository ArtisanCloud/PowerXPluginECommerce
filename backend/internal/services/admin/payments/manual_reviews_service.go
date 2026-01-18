package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrManualReviewServiceUnavailable = errors.New("manual payment review service unavailable")
	ErrManualReviewNotFound           = errors.New("manual payment review not found")
	ErrManualReviewInvalidStatus      = errors.New("manual payment review status invalid")
	ErrManualReviewSameOperator       = errors.New("reviewer must be different from submitter")
	ErrManualReviewReasonRequired     = errors.New("review reason is required")
	ErrManualReviewOrderNotPayable    = errors.New("order is not payable")
)

type ManualReviewService struct {
	deps            *app.Deps
	reviewRepo      *paymentrepo.PaymentManualReviewRepository
	transactionRepo *paymentrepo.PaymentTransactionRepository
	orderRepo       *orderrepo.OrderRepository
	eventRepo       *orderrepo.OrderEventRepository
	logger          *paymentslogger.Logger
}

func NewManualReviewService(deps *app.Deps) *ManualReviewService {
	if deps == nil || deps.DB == nil {
		return &ManualReviewService{deps: deps}
	}
	var obsLogger *paymentslogger.Logger
	if entry := deps.RuntimeLogger(deps.Ctx, "payments", nil); entry != nil {
		obsLogger = paymentslogger.NewLogger(entry)
	}
	return &ManualReviewService{
		deps:            deps,
		reviewRepo:      paymentrepo.NewPaymentManualReviewRepository(deps.DB),
		transactionRepo: paymentrepo.NewPaymentTransactionRepository(deps.DB),
		orderRepo:       orderrepo.NewOrderRepository(deps.DB),
		eventRepo:       orderrepo.NewOrderEventRepository(deps.DB),
		logger:          obsLogger,
	}
}

func (s *ManualReviewService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.reviewRepo != nil && s.transactionRepo != nil && s.orderRepo != nil && s.eventRepo != nil
}

func (s *ManualReviewService) ListReviews(ctx context.Context, tenantUUID, adminID, orderID string) ([]ManualPaymentReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrManualReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	rows, err := s.reviewRepo.ListByOrderID(ctx, tenantUUID, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]ManualPaymentReviewDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toManualReviewDTO(&row))
	}
	return out, nil
}

func (s *ManualReviewService) CreateReview(ctx context.Context, tenantUUID, adminID string, req ManualPaymentCreateRequest) (*ManualPaymentReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrManualReviewServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	adminID = strings.TrimSpace(adminID)
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.Currency = strings.TrimSpace(req.Currency)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.ProofNo = strings.TrimSpace(req.ProofNo)
	req.Note = strings.TrimSpace(req.Note)
	if adminID == "" {
		return nil, errors.New("admin id is required")
	}
	if req.OrderID == "" {
		return nil, errors.New("order id is required")
	}
	if req.AmountMinor <= 0 {
		return nil, errors.New("amount is required")
	}
	if req.PayMethod == "" {
		return nil, errors.New("pay method is required")
	}
	var created *models.PaymentManualReview
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		orderRow, err := s.orderRepo.LockByID(ctx, tx, tenantUUID, req.OrderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrManualReviewOrderNotPayable
			}
			return err
		}
		if strings.TrimSpace(orderRow.Status) != "pending_payment" {
			return ErrManualReviewOrderNotPayable
		}
		currency := req.Currency
		if currency == "" {
			currency = orderRow.Currency
		}
		if currency != orderRow.Currency {
			return errors.New("currency mismatch")
		}
		if req.AmountMinor != orderRow.TotalAmount {
			return errors.New("amount must equal order total")
		}
		now := time.Now().UTC()
		row := &models.PaymentManualReview{
			BaseModel:   models.BaseModel{TenantUuid: tenantUUID},
			OrderID:     req.OrderID,
			OrderNo:     orderRow.OrderNo,
			ProviderID:  req.ProviderID,
			PayMethod:   req.PayMethod,
			AmountMinor: req.AmountMinor,
			Currency:    currency,
			Status:      "pending_review",
			SubmittedBy: adminID,
			SubmittedAt: now,
			ProofNo:     req.ProofNo,
			Note:        req.Note,
		}
		if err := tx.WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.logger != nil {
		requestID, _ := authx.RequestIDFromContext(ctx)
		s.logger.EmitEvent(paymentslogger.Event{
			Action:    "manual_review.created",
			TenantID:  tenantUUID,
			ActorID:   adminID,
			RequestID: requestID,
			OrderID:   created.OrderID,
			OrderNo:   created.OrderNo,
			Amount:    created.AmountMinor,
			Currency:  created.Currency,
			Status:    created.Status,
			Metadata: map[string]any{
				"review_id":  created.ID,
				"pay_method": created.PayMethod,
				"provider_id": created.ProviderID,
				"proof_no":   created.ProofNo,
			},
		})
	}
	resp := toManualReviewDTO(created)
	return &resp, nil
}

func (s *ManualReviewService) ApproveReview(ctx context.Context, tenantUUID, adminID string, reviewID uint64, req ManualPaymentReviewRequest) (*ManualPaymentReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrManualReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if reviewID == 0 {
		return nil, errors.New("review id is required")
	}
	var updated *models.PaymentManualReview
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		row, err := s.reviewRepo.LockByID(ctx, tx, tenantUUID, reviewID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrManualReviewNotFound
			}
			return err
		}
		if strings.TrimSpace(row.Status) != "pending_review" {
			return ErrManualReviewInvalidStatus
		}
		if row.SubmittedBy == adminID {
			return ErrManualReviewSameOperator
		}
		orderRow, err := s.orderRepo.LockByID(ctx, tx, tenantUUID, row.OrderID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(orderRow.Status) != "pending_payment" {
			return ErrManualReviewOrderNotPayable
		}
		now := time.Now().UTC()
		transactionNo := fmt.Sprintf("MN-%d-%d", reviewID, now.Unix())
		payTx := &models.PaymentTransaction{
			BaseModel:      models.BaseModel{TenantUuid: tenantUUID},
			TransactionNo:  transactionNo,
			OrderID:        row.OrderID,
			OrderNo:        row.OrderNo,
			ProviderID:     row.ProviderID,
			PayMethod:      row.PayMethod,
			AmountTotal:    row.AmountMinor,
			AmountCurrency: row.Currency,
			FeeAmount:      0,
			Status:         "paid",
			CompletedAt:    &now,
		}
		if err := tx.WithContext(ctx).Create(payTx).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"status":        "approved",
			"reviewed_by":   adminID,
			"reviewed_at":   now,
			"review_reason": strings.TrimSpace(req.Reason),
			"transaction_id": payTx.ID,
		}
		if err := tx.WithContext(ctx).
			Model(&models.PaymentManualReview{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, reviewID).
			Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Model(&ordermodel.Order{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, row.OrderID).
			Updates(map[string]any{
				"status":     "paid",
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		eventPayload, _ := jsonManualReviewPayload(ctx, row, payTx, adminID, "approved")
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      row.OrderID,
			EventType:    "order.paid.manual",
			OperatorType: "admin",
			Operator:     adminID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, reviewID).First(&row).Error; err != nil {
			return err
		}
		updated = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.logger != nil {
		requestID, _ := authx.RequestIDFromContext(ctx)
		s.logger.EmitEvent(paymentslogger.Event{
			Action:        "manual_review.approved",
			TenantID:      tenantUUID,
			ActorID:       adminID,
			RequestID:     requestID,
			OrderID:       updated.OrderID,
			OrderNo:       updated.OrderNo,
			TransactionID: fmt.Sprintf("%d", updated.TransactionID),
			ProviderID:    fmt.Sprintf("%d", updated.ProviderID),
			Amount:        updated.AmountMinor,
			Currency:      updated.Currency,
			Status:        updated.Status,
			Result:        "approved",
			Reason:        strings.TrimSpace(req.Reason),
			Metadata: map[string]any{
				"review_id":  updated.ID,
				"pay_method": updated.PayMethod,
				"proof_no":   updated.ProofNo,
			},
		})
	}
	resp := toManualReviewDTO(updated)
	return &resp, nil
}

func (s *ManualReviewService) RejectReview(ctx context.Context, tenantUUID, adminID string, reviewID uint64, req ManualPaymentReviewRequest) (*ManualPaymentReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrManualReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if reviewID == 0 {
		return nil, errors.New("review id is required")
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, ErrManualReviewReasonRequired
	}
	var updated *models.PaymentManualReview
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		row, err := s.reviewRepo.LockByID(ctx, tx, tenantUUID, reviewID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrManualReviewNotFound
			}
			return err
		}
		if strings.TrimSpace(row.Status) != "pending_review" {
			return ErrManualReviewInvalidStatus
		}
		if row.SubmittedBy == adminID {
			return ErrManualReviewSameOperator
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":        "rejected",
			"reviewed_by":   adminID,
			"reviewed_at":   now,
			"review_reason": reason,
		}
		if err := tx.WithContext(ctx).
			Model(&models.PaymentManualReview{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, reviewID).
			Updates(updates).Error; err != nil {
			return err
		}
		eventPayload, _ := jsonManualReviewPayload(ctx, row, nil, adminID, "rejected")
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      row.OrderID,
			EventType:    "order.manual_payment.rejected",
			OperatorType: "admin",
			Operator:     adminID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, reviewID).First(&row).Error; err != nil {
			return err
		}
		updated = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.logger != nil {
		requestID, _ := authx.RequestIDFromContext(ctx)
		s.logger.EmitEvent(paymentslogger.Event{
			Action:    "manual_review.rejected",
			TenantID:  tenantUUID,
			ActorID:   adminID,
			RequestID: requestID,
			OrderID:   updated.OrderID,
			OrderNo:   updated.OrderNo,
			ProviderID: fmt.Sprintf("%d", updated.ProviderID),
			Amount:    updated.AmountMinor,
			Currency:  updated.Currency,
			Status:    updated.Status,
			Result:    "rejected",
			Reason:    updated.ReviewReason,
			Metadata: map[string]any{
				"review_id":  updated.ID,
				"pay_method": updated.PayMethod,
				"proof_no":   updated.ProofNo,
			},
		})
	}
	resp := toManualReviewDTO(updated)
	return &resp, nil
}

func toManualReviewDTO(row *models.PaymentManualReview) ManualPaymentReviewDTO {
	if row == nil {
		return ManualPaymentReviewDTO{}
	}
	return ManualPaymentReviewDTO{
		ID:            row.ID,
		OrderID:       row.OrderID,
		OrderNo:       row.OrderNo,
		TransactionID: row.TransactionID,
		ProviderID:    row.ProviderID,
		PayMethod:     row.PayMethod,
		AmountMinor:   row.AmountMinor,
		Currency:      row.Currency,
		Status:        row.Status,
		SubmittedBy:   row.SubmittedBy,
		SubmittedAt:   row.SubmittedAt,
		ReviewedBy:    row.ReviewedBy,
		ReviewedAt:    row.ReviewedAt,
		ReviewReason:  row.ReviewReason,
		ProofNo:       row.ProofNo,
		Note:          row.Note,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func jsonManualReviewPayload(ctx context.Context, review *models.PaymentManualReview, tx *models.PaymentTransaction, reviewer, action string) ([]byte, error) {
	requestID, _ := authx.RequestIDFromContext(ctx)
	var txID uint64
	var txNo string
	if tx != nil {
		txID = tx.ID
		txNo = tx.TransactionNo
	}
	payload := map[string]any{
		"requestId":       requestID,
		"manualReviewId":  review.ID,
		"orderId":         review.OrderID,
		"orderNo":         review.OrderNo,
		"amount":          review.AmountMinor,
		"currency":        review.Currency,
		"payMethod":       review.PayMethod,
		"providerId":      review.ProviderID,
		"proofNo":         review.ProofNo,
		"note":            review.Note,
		"reviewer":        reviewer,
		"action":          action,
		"transactionId":   txID,
		"transactionNo":   txNo,
	}
	return json.Marshal(payload)
}
