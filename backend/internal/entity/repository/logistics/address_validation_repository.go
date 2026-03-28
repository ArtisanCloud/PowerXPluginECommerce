package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type AddressValidationRepository struct {
	*Repository[LogisticsModel.AddressValidation]
}

func NewAddressValidationRepository(db *gorm.DB) *AddressValidationRepository {
	return &AddressValidationRepository{Repository: NewRepository[LogisticsModel.AddressValidation](db)}
}

func (r *AddressValidationRepository) Create(ctx context.Context, row *LogisticsModel.AddressValidation) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *AddressValidationRepository) GetByRequestKey(ctx context.Context, key string) (*LogisticsModel.AddressValidation, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.AddressValidation
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(key)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AddressValidationRepository) List(ctx context.Context, waybillNo string, limit int) ([]LogisticsModel.AddressValidation, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillNo) != "" {
		db = db.Where("waybill_no = ?", strings.TrimSpace(waybillNo))
	}
	var rows []LogisticsModel.AddressValidation
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
