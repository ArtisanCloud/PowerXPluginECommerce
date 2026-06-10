package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// TemplateRepository handles coupon template persistence.
type TemplateRepository struct {
	*repository.BaseRepository[couponmodel.CouponTemplate]
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{BaseRepository: repository.NewBaseRepository[couponmodel.CouponTemplate](db)}
}

func (r *TemplateRepository) ListActiveByIDs(ctx context.Context, tenantUUID string, ids []string, at time.Time) ([]couponmodel.CouponTemplate, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if len(ids) == 0 {
		return []couponmodel.CouponTemplate{}, nil
	}
	trimmed := make([]string, 0, len(ids))
	for _, id := range ids {
		if v := strings.TrimSpace(id); v != "" {
			trimmed = append(trimmed, v)
		}
	}
	if len(trimmed) == 0 {
		return []couponmodel.CouponTemplate{}, nil
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	var rows []couponmodel.CouponTemplate
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id IN ?", tenantUUID, trimmed).
		Where("status = ?", "active").
		Where("valid_from <= ? AND valid_to >= ?", at, at).
		Find(&rows).Error
	return rows, err
}

func (r *TemplateRepository) MapByID(ctx context.Context, tenantUUID string, ids []string, at time.Time) (map[string]couponmodel.CouponTemplate, error) {
	list, err := r.ListActiveByIDs(ctx, tenantUUID, ids, at)
	if err != nil {
		return nil, err
	}
	out := make(map[string]couponmodel.CouponTemplate, len(list))
	for _, row := range list {
		out[strings.TrimSpace(row.ID)] = row
	}
	return out, nil
}

func (r *TemplateRepository) MapByIDAnyWithTx(ctx context.Context, tx *gorm.DB, tenantUUID string, ids []string) (map[string]couponmodel.CouponTemplate, error) {
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
		return map[string]couponmodel.CouponTemplate{}, nil
	}
	var rows []couponmodel.CouponTemplate
	err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND id IN ?", tenantUUID, trimmed).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]couponmodel.CouponTemplate, len(rows))
	for _, row := range rows {
		out[strings.TrimSpace(row.ID)] = row
	}
	return out, nil
}
