package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// CarrierRepository provides tenant-scoped carrier persistence operations.
type CarrierRepository struct {
	*Repository[LogisticsModel.Carrier]
}

func NewCarrierRepository(db *gorm.DB) *CarrierRepository {
	return &CarrierRepository{Repository: NewRepository[LogisticsModel.Carrier](db)}
}

func (r *CarrierRepository) List(ctx context.Context) ([]LogisticsModel.Carrier, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.Carrier
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *CarrierRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.Carrier, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.Carrier
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CarrierRepository) GetByCode(ctx context.Context, code string) (*LogisticsModel.Carrier, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.Carrier
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND code = ?", tenantUUID, strings.TrimSpace(code)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CarrierRepository) Create(ctx context.Context, carrier *LogisticsModel.Carrier) error {
	if carrier == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(carrier.TenantUUID) == "" {
		carrier.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(carrier).Error
}

func (r *CarrierRepository) Save(ctx context.Context, carrier *LogisticsModel.Carrier) error {
	if carrier == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(carrier.TenantUUID) == "" {
		carrier.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Save(carrier).Error
}

// CarrierServiceRepository handles carrier service rows.
type CarrierServiceRepository struct {
	*Repository[LogisticsModel.CarrierService]
}

func NewCarrierServiceRepository(db *gorm.DB) *CarrierServiceRepository {
	return &CarrierServiceRepository{Repository: NewRepository[LogisticsModel.CarrierService](db)}
}

func (r *CarrierServiceRepository) ListByCarrierID(ctx context.Context, carrierID string) ([]LogisticsModel.CarrierService, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.CarrierService
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND carrier_id = ?", tenantUUID, strings.TrimSpace(carrierID)).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *CarrierServiceRepository) Create(ctx context.Context, service *LogisticsModel.CarrierService) error {
	if service == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(service.TenantUUID) == "" {
		service.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(service).Error
}
