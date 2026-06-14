package promotion

import (
	"context"
	"strings"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderSnapshotRepository handles promotion order snapshot persistence.
type OrderSnapshotRepository struct {
	*repository.BaseRepository[promotionmodel.OrderSnapshot]
}

func NewOrderSnapshotRepository(db *gorm.DB) *OrderSnapshotRepository {
	return &OrderSnapshotRepository{BaseRepository: repository.NewBaseRepository[promotionmodel.OrderSnapshot](db)}
}

func (r *OrderSnapshotRepository) FindByOrderID(ctx context.Context, tenantUUID, orderID string) (*promotionmodel.OrderSnapshot, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	var row promotionmodel.OrderSnapshot
	if err := r.DB.WithContext(ctx).Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderSnapshotRepository) UpsertWithTx(ctx context.Context, tx *gorm.DB, row *promotionmodel.OrderSnapshot) error {
	if r == nil || r.DB == nil {
		return gorm.ErrInvalidDB
	}
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	if row == nil {
		return gorm.ErrInvalidData
	}
	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tenant_uuid"}, {Name: "order_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"currency", "base_total_minor", "promotion_discount_minor", "after_promotion_total_minor",
			"applied_promotions", "rejected_promotions", "line_allocations", "priced_at", "updated_at",
		}),
	}).Create(row).Error
}
