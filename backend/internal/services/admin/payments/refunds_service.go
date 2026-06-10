package payments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

var (
	ErrRefundServiceUnavailable = errors.New("payment refund service unavailable")
	ErrRefundAmountRequired     = errors.New("refund amount is required")
)

type RefundService struct {
	deps            *app.Deps
	refundRepo      *paymentrepo.PaymentRefundRepository
	transactionRepo *paymentrepo.PaymentTransactionRepository
	couponRefundSvc *couponsvc.RefundService
	logger          *paymentslogger.Logger
}

func NewRefundService(deps *app.Deps) *RefundService {
	if deps == nil || deps.DB == nil {
		return &RefundService{deps: deps}
	}
	var obsLogger *paymentslogger.Logger
	if entry := deps.RuntimeLogger(deps.Ctx, "payments", nil); entry != nil {
		obsLogger = paymentslogger.NewLogger(entry)
	}
	return &RefundService{
		deps:            deps,
		refundRepo:      paymentrepo.NewPaymentRefundRepository(deps.DB),
		transactionRepo: paymentrepo.NewPaymentTransactionRepository(deps.DB),
		couponRefundSvc: couponsvc.NewRefundService(deps),
		logger:          obsLogger,
	}
}

func (s *RefundService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.refundRepo != nil && s.transactionRepo != nil
}

func (s *RefundService) CreateRefund(ctx context.Context, tenantUUID, adminID string, transactionID uint64, req CreateRefundRequest) (*RefundDTO, error) {
	if !s.Ready() {
		return nil, ErrRefundServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if req.AmountMinor <= 0 {
		return nil, ErrRefundAmountRequired
	}
	var txRow models.PaymentTransaction
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, transactionID).First(&txRow).Error; err != nil {
		return nil, err
	}

	refundNo := fmt.Sprintf("RF-%d-%d", transactionID, time.Now().Unix())
	row := &models.PaymentRefund{
		BaseModel:      models.BaseModel{TenantUuid: strings.TrimSpace(tenantUUID)},
		TransactionID:  transactionID,
		RefundNo:       refundNo,
		RefundAmount:   req.AmountMinor,
		RefundCurrency: txRow.AmountCurrency,
		Status:         "requested",
		Reason:         strings.TrimSpace(req.Reason),
	}
	var created *models.PaymentRefund
	err := s.refundRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
		if s.couponRefundSvc != nil && s.couponRefundSvc.Ready() && strings.TrimSpace(txRow.OrderID) != "" {
			if _, err := s.couponRefundSvc.RefundWithTx(ctx, tx, couponsvc.RefundInput{
				TenantUUID: tenantUUID,
				OrderID:    txRow.OrderID,
				RefundNo:   refundNo,
				Operator:   adminID,
				Reason:     "payment_refund",
			}); err != nil {
				return err
			}
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
			Action:        "refund.created",
			TenantID:      tenantUUID,
			ActorID:       adminID,
			RequestID:     requestID,
			OrderID:       txRow.OrderID,
			OrderNo:       txRow.OrderNo,
			TransactionID: fmt.Sprintf("%d", transactionID),
			ProviderID:    fmt.Sprintf("%d", txRow.ProviderID),
			Amount:        created.RefundAmount,
			Currency:      created.RefundCurrency,
			Status:        created.Status,
			Reason:        strings.TrimSpace(req.Reason),
			Metadata: map[string]any{
				"refund_id": created.ID,
				"refund_no": created.RefundNo,
			},
		})
	}
	return &RefundDTO{
		ID:           created.ID,
		RefundNo:     created.RefundNo,
		RefundAmount: created.RefundAmount,
		Currency:     created.RefundCurrency,
		Status:       created.Status,
		CreatedAt:    created.CreatedAt,
		CompletedAt:  created.CompletedAt,
	}, nil
}
