package after_sales

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

var (
	ErrPaymentGuardUnavailable = errors.New("after-sales payment guard unavailable")
	ErrRefundAlreadyApplied    = errors.New("refund already applied for this order item")
	ErrRefundAlreadyInPayment  = errors.New("refund already exists in payment records")
)

type PaymentGuardService struct {
	deps *app.Deps
}

func NewPaymentGuardService(deps *app.Deps) *PaymentGuardService {
	return &PaymentGuardService{deps: deps}
}

func (s *PaymentGuardService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil
}

func (s *PaymentGuardService) EnsureRefundAllowed(ctx context.Context, tenantUUID string, row *AfterSalesModel.AfterSaleCase) error {
	if !s.Ready() {
		return ErrPaymentGuardUnavailable
	}
	if row == nil {
		return gorm.ErrInvalidData
	}
	caseType := strings.TrimSpace(strings.ToLower(row.CaseType))
	if caseType != "refund_only" && caseType != "return_refund" {
		return nil
	}
	tenantUUID = strings.TrimSpace(tenantUUID)

	var duplicateCount int64
	caseTable := models.S(models.TableAfterSalesCases)
	if err := s.deps.DB.WithContext(ctx).
		Table(caseTable).
		Where("tenant_uuid = ? AND order_item_id = ? AND id <> ? AND case_type IN ? AND status IN ?",
			tenantUUID,
			strings.TrimSpace(row.OrderItemID),
			strings.TrimSpace(row.ID),
			[]string{"refund_only", "return_refund"},
			[]string{CaseStatusApproved, CaseStatusCompleted},
		).
		Count(&duplicateCount).Error; err != nil {
		return err
	}
	if duplicateCount > 0 {
		return ErrRefundAlreadyApplied
	}

	refundTable := models.S(models.TablePaymentRefunds)
	txTable := models.S(models.TablePaymentTransactions)
	if err := s.deps.DB.WithContext(ctx).
		Table(refundTable+" AS pr").
		Joins("JOIN "+txTable+" AS pt ON pt.id = pr.transaction_id AND pt.tenant_uuid = pr.tenant_uuid").
		Where("pr.tenant_uuid = ? AND pt.order_id = ? AND pr.status NOT IN ?",
			tenantUUID,
			strings.TrimSpace(row.OrderID),
			[]string{"failed", "cancelled", "rejected", "closed"},
		).
		Count(&duplicateCount).Error; err != nil {
		return err
	}
	if duplicateCount > 0 {
		return ErrRefundAlreadyInPayment
	}
	return nil
}
