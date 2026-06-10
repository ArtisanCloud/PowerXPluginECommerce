package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AssetRepository handles coupon asset persistence.
type AssetRepository struct {
	*repository.BaseRepository[couponmodel.CouponAsset]
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{BaseRepository: repository.NewBaseRepository[couponmodel.CouponAsset](db)}
}

func (r *AssetRepository) ListAvailableByUserAndIDs(ctx context.Context, tenantUUID, userID string, ids []string, at time.Time) ([]couponmodel.CouponAsset, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	userID = strings.TrimSpace(userID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	trimmed := make([]string, 0, len(ids))
	for _, id := range ids {
		if v := strings.TrimSpace(id); v != "" {
			trimmed = append(trimmed, v)
		}
	}
	if len(trimmed) == 0 {
		return []couponmodel.CouponAsset{}, nil
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}

	query := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND user_id = ? AND id IN ?", tenantUUID, userID, trimmed).
		Where("status = ?", "available")
	query = query.Where("(valid_from IS NULL OR valid_from <= ?) AND (valid_to IS NULL OR valid_to >= ?)", at, at)

	var rows []couponmodel.CouponAsset
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AssetRepository) LockByIDsForUpdate(ctx context.Context, tx *gorm.DB, tenantUUID string, ids []string) ([]couponmodel.CouponAsset, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	trimmed := make([]string, 0, len(ids))
	for _, id := range ids {
		if v := strings.TrimSpace(id); v != "" {
			trimmed = append(trimmed, v)
		}
	}
	if len(trimmed) == 0 {
		return []couponmodel.CouponAsset{}, nil
	}

	query := tx.WithContext(ctx).Where("tenant_uuid = ? AND id IN ?", tenantUUID, trimmed)
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var rows []couponmodel.CouponAsset
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AssetRepository) SaveWithTx(ctx context.Context, tx *gorm.DB, asset *couponmodel.CouponAsset) error {
	if r == nil || r.DB == nil {
		return gorm.ErrInvalidDB
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if asset == nil {
		return errors.New("asset is required")
	}
	return tx.WithContext(ctx).Save(asset).Error
}

func (r *AssetRepository) LockReservedByOrderForUpdate(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) ([]couponmodel.CouponAsset, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	query := tx.WithContext(ctx).Where("tenant_uuid = ? AND reserved_order_id = ? AND status = ?", tenantUUID, orderID, "reserved")
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var rows []couponmodel.CouponAsset
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AssetRepository) LockRedeemedByOrderForUpdate(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) ([]couponmodel.CouponAsset, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	query := tx.WithContext(ctx).Where("tenant_uuid = ? AND reserved_order_id = ? AND status = ?", tenantUUID, orderID, "redeemed")
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var rows []couponmodel.CouponAsset
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
