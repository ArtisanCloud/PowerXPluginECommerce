package coupon

import (
	"context"
	"errors"
	"strings"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderSnapshotRepository handles order coupon snapshots.
type OrderSnapshotRepository struct {
	*repository.BaseRepository[couponmodel.OrderCouponSnapshot]
}

func NewOrderSnapshotRepository(db *gorm.DB) *OrderSnapshotRepository {
	return &OrderSnapshotRepository{BaseRepository: repository.NewBaseRepository[couponmodel.OrderCouponSnapshot](db)}
}

func (r *OrderSnapshotRepository) UpsertWithTx(ctx context.Context, tx *gorm.DB, row *couponmodel.OrderCouponSnapshot) error {
	if r == nil || r.DB == nil {
		return gorm.ErrInvalidDB
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if row == nil {
		return errors.New("snapshot row is required")
	}
	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_uuid"}, {Name: "order_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"currency", "base_total_minor", "discount_total_minor", "payable_total_minor",
			"line_allocations", "applied_coupons", "rejected_coupons", "priced_at", "updated_at",
		}),
	}).Create(row).Error
}

func (r *OrderSnapshotRepository) GetByOrderID(ctx context.Context, tenantUUID, orderID string) (*couponmodel.OrderCouponSnapshot, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var row couponmodel.OrderCouponSnapshot
	if err := r.DB.WithContext(ctx).Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
