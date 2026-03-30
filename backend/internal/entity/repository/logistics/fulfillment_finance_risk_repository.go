package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type FinanceRiskFilter struct {
	CarrierID string
	Status    string
	RiskLevel string
	Limit     int
}

type FulfillmentFinanceRiskRepository struct {
	*Repository[LogisticsModel.FulfillmentFinanceRisk]
}

func NewFulfillmentFinanceRiskRepository(db *gorm.DB) *FulfillmentFinanceRiskRepository {
	return &FulfillmentFinanceRiskRepository{Repository: NewRepository[LogisticsModel.FulfillmentFinanceRisk](db)}
}

func (r *FulfillmentFinanceRiskRepository) Save(ctx context.Context, row *LogisticsModel.FulfillmentFinanceRisk) error {
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
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *FulfillmentFinanceRiskRepository) List(ctx context.Context, filter FinanceRiskFilter) ([]LogisticsModel.FulfillmentFinanceRisk, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.CarrierID); v != "" {
		q = q.Where("carrier_id = ?", v)
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		q = q.Where("status = ?", v)
	}
	if v := strings.TrimSpace(filter.RiskLevel); v != "" {
		q = q.Where("risk_level = ?", v)
	}
	var rows []LogisticsModel.FulfillmentFinanceRisk
	err = q.Order("composite_risk_score DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FulfillmentFinanceRiskRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.FulfillmentFinanceRisk, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentFinanceRisk
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FulfillmentFinanceRiskRepository) GetByWaybillID(ctx context.Context, waybillID string) (*LogisticsModel.FulfillmentFinanceRisk, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentFinanceRisk
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type FulfillmentFinanceRiskAuditRepository struct {
	*Repository[LogisticsModel.FulfillmentFinanceRiskAudit]
}

func NewFulfillmentFinanceRiskAuditRepository(db *gorm.DB) *FulfillmentFinanceRiskAuditRepository {
	return &FulfillmentFinanceRiskAuditRepository{Repository: NewRepository[LogisticsModel.FulfillmentFinanceRiskAudit](db)}
}

func (r *FulfillmentFinanceRiskAuditRepository) Save(ctx context.Context, row *LogisticsModel.FulfillmentFinanceRiskAudit) error {
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
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *FulfillmentFinanceRiskAuditRepository) List(ctx context.Context, riskID string, limit int) ([]LogisticsModel.FulfillmentFinanceRiskAudit, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(riskID); v != "" {
		q = q.Where("risk_id = ?", v)
	}
	var rows []LogisticsModel.FulfillmentFinanceRiskAudit
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FulfillmentFinanceRiskAuditRepository) GetByRequestKey(ctx context.Context, riskID, requestKey string) (*LogisticsModel.FulfillmentFinanceRiskAudit, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.FulfillmentFinanceRiskAudit
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND risk_id = ? AND request_key = ?", tenantUUID, strings.TrimSpace(riskID), strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
