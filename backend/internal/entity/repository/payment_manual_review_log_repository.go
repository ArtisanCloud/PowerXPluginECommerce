package repository

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// PaymentManualReviewLogRepository handles manual review log records.
type PaymentManualReviewLogRepository struct {
	*BaseRepository[models.PaymentManualReviewLog]
}

func NewPaymentManualReviewLogRepository(db *gorm.DB) *PaymentManualReviewLogRepository {
	return &PaymentManualReviewLogRepository{BaseRepository: NewBaseRepository[models.PaymentManualReviewLog](db)}
}

func (r *PaymentManualReviewLogRepository) ListByOrderID(ctx context.Context, tenantUUID, orderID string) ([]models.PaymentManualReviewLog, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("payment manual review log repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var rows []models.PaymentManualReviewLog
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).
		Order("created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
