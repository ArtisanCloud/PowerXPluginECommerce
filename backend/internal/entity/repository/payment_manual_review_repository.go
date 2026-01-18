package repository

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaymentManualReviewRepository handles manual payment review records.
type PaymentManualReviewRepository struct {
	*BaseRepository[models.PaymentManualReview]
}

func NewPaymentManualReviewRepository(db *gorm.DB) *PaymentManualReviewRepository {
	return &PaymentManualReviewRepository{BaseRepository: NewBaseRepository[models.PaymentManualReview](db)}
}

func (r *PaymentManualReviewRepository) LockByID(ctx context.Context, tx *gorm.DB, tenantUUID string, id uint64) (*models.PaymentManualReview, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("manual review repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, ErrTenantUuidRequired
	}
	if id == 0 {
		return nil, errors.New("review id is required")
	}
	var row models.PaymentManualReview
	query := tx.WithContext(ctx)
	if query.Dialector != nil && query.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaymentManualReviewRepository) ListByOrderID(ctx context.Context, tenantUUID, orderID string) ([]models.PaymentManualReview, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("manual review repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, ErrTenantUuidRequired
	}
	query := r.DB.WithContext(ctx).Model(&models.PaymentManualReview{}).Where("tenant_uuid = ?", tenantUUID)
	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}
	var rows []models.PaymentManualReview
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
