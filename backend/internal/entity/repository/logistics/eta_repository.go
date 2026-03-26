package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// ETAPolicyRepository manages eta policy persistence.
type ETAPolicyRepository struct {
	*Repository[LogisticsModel.ETAPolicy]
}

func NewETAPolicyRepository(db *gorm.DB) *ETAPolicyRepository {
	return &ETAPolicyRepository{Repository: NewRepository[LogisticsModel.ETAPolicy](db)}
}

func (r *ETAPolicyRepository) FindByCarrierService(ctx context.Context, carrierID, serviceCode, zone string) (*LogisticsModel.ETAPolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	carrierID = strings.TrimSpace(carrierID)
	serviceCode = strings.TrimSpace(serviceCode)
	if carrierID == "" || serviceCode == "" {
		return nil, gorm.ErrRecordNotFound
	}
	query := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND carrier_id = ? AND service_code = ?", tenantUUID, carrierID, serviceCode)
	zone = strings.ToUpper(strings.TrimSpace(zone))
	if zone != "" {
		var row LogisticsModel.ETAPolicy
		err := query.
			Where("destination_zone = ?", zone).
			Order("updated_at DESC").
			First(&row).Error
		if err == nil {
			return &row, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	var fallback LogisticsModel.ETAPolicy
	err = query.
		Where("destination_zone = ?", "GLOBAL").
		Order("updated_at DESC").
		First(&fallback).Error
	if err != nil {
		return nil, err
	}
	return &fallback, nil
}

// ETARecordRepository manages computed eta snapshots.
type ETARecordRepository struct {
	*Repository[LogisticsModel.ETARecord]
}

func NewETARecordRepository(db *gorm.DB) *ETARecordRepository {
	return &ETARecordRepository{Repository: NewRepository[LogisticsModel.ETARecord](db)}
}

func (r *ETARecordRepository) GetByWaybillID(ctx context.Context, waybillID string) (*LogisticsModel.ETARecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ETARecord
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ETARecordRepository) ListByWaybillIDs(ctx context.Context, waybillIDs []string) ([]LogisticsModel.ETARecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(waybillIDs))
	for _, id := range waybillIDs {
		if v := strings.TrimSpace(id); v != "" {
			ids = append(ids, v)
		}
	}
	if len(ids) == 0 {
		return []LogisticsModel.ETARecord{}, nil
	}
	var rows []LogisticsModel.ETARecord
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id IN ?", tenantUUID, ids).
		Order("updated_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *ETARecordRepository) UpsertByWaybillID(ctx context.Context, row *LogisticsModel.ETARecord) error {
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
	if row.LastComputedAt.IsZero() {
		row.LastComputedAt = time.Now().UTC()
	}
	existing, err := r.GetByWaybillID(ctx, row.WaybillID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.DB.WithContext(ctx).Create(row).Error
		}
		return err
	}
	row.ID = existing.ID
	return r.DB.WithContext(ctx).Save(row).Error
}
