package repository

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// OrderBenefitReviewRepository handles benefit review records.
type OrderBenefitReviewRepository struct {
	*BaseRepository[models.OrderBenefitReview]
}

func NewOrderBenefitReviewRepository(db *gorm.DB) *OrderBenefitReviewRepository {
	return &OrderBenefitReviewRepository{BaseRepository: NewBaseRepository[models.OrderBenefitReview](db)}
}

func (r *OrderBenefitReviewRepository) LockByID(ctx context.Context, tx *gorm.DB, tenantUUID string, id uint64) (*models.OrderBenefitReview, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order benefit review repository is not initialized")
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
	var row models.OrderBenefitReview
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderBenefitReviewRepository) ListByOrderID(ctx context.Context, tenantUUID, orderID string) ([]models.OrderBenefitReview, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order benefit review repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var rows []models.OrderBenefitReview
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).
		Order("created_at DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *OrderBenefitReviewRepository) CountActiveByOrder(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) (int64, error) {
	if r == nil || r.DB == nil {
		return 0, errors.New("order benefit review repository is not initialized")
	}
	if tx == nil {
		return 0, errors.New("transaction is required")
	}
	var count int64
	if err := tx.WithContext(ctx).
		Model(&models.OrderBenefitReview{}).
		Where("tenant_uuid = ? AND order_id = ? AND status IN ('pending_review','approved')", tenantUUID, orderID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *OrderBenefitReviewRepository) CountApprovedByCode(ctx context.Context, tx *gorm.DB, tenantUUID, benefitType, benefitCode string, excludeID uint64) (int64, error) {
	if r == nil || r.DB == nil {
		return 0, errors.New("order benefit review repository is not initialized")
	}
	if tx == nil {
		return 0, errors.New("transaction is required")
	}
	query := tx.WithContext(ctx).
		Model(&models.OrderBenefitReview{}).
		Where("tenant_uuid = ? AND benefit_type = ? AND benefit_code = ? AND status = 'approved'", tenantUUID, benefitType, benefitCode)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
