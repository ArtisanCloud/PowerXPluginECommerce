package coupon

import (
	"context"
	"errors"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UsageLogRepository handles coupon action logs.
type UsageLogRepository struct {
	*repository.BaseRepository[couponmodel.CouponUsageLog]
}

func NewUsageLogRepository(db *gorm.DB) *UsageLogRepository {
	return &UsageLogRepository{BaseRepository: repository.NewBaseRepository[couponmodel.CouponUsageLog](db)}
}

func (r *UsageLogRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, row *couponmodel.CouponUsageLog) error {
	if r == nil || r.DB == nil {
		return gorm.ErrInvalidDB
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if row == nil {
		return errors.New("usage log row is required")
	}
	return tx.WithContext(ctx).Create(row).Error
}

func (r *UsageLogRepository) CreateWithTxIdempotent(ctx context.Context, tx *gorm.DB, row *couponmodel.CouponUsageLog) (bool, error) {
	if r == nil || r.DB == nil {
		return false, gorm.ErrInvalidDB
	}
	if tx == nil {
		return false, errors.New("transaction is required")
	}
	if row == nil {
		return false, errors.New("usage log row is required")
	}
	res := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(row)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
