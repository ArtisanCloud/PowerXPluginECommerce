package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type CrossborderDocumentRepository struct {
	*Repository[LogisticsModel.CrossborderDocument]
}

func NewCrossborderDocumentRepository(db *gorm.DB) *CrossborderDocumentRepository {
	return &CrossborderDocumentRepository{Repository: NewRepository[LogisticsModel.CrossborderDocument](db)}
}

func (r *CrossborderDocumentRepository) Save(ctx context.Context, row *LogisticsModel.CrossborderDocument) error {
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

func (r *CrossborderDocumentRepository) List(ctx context.Context, waybillNo string, limit int) ([]LogisticsModel.CrossborderDocument, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillNo) != "" {
		q = q.Where("waybill_no = ?", strings.TrimSpace(waybillNo))
	}
	var rows []LogisticsModel.CrossborderDocument
	err = q.Order("updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CrossborderDocumentRepository) GetByUniq(ctx context.Context, waybillNo, docType string) (*LogisticsModel.CrossborderDocument, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CrossborderDocument
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_no = ? AND doc_type = ?", tenantUUID, strings.TrimSpace(waybillNo), strings.TrimSpace(docType)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type CrossborderTaxQuoteRepository struct {
	*Repository[LogisticsModel.CrossborderTaxQuote]
}

func NewCrossborderTaxQuoteRepository(db *gorm.DB) *CrossborderTaxQuoteRepository {
	return &CrossborderTaxQuoteRepository{Repository: NewRepository[LogisticsModel.CrossborderTaxQuote](db)}
}

func (r *CrossborderTaxQuoteRepository) Save(ctx context.Context, row *LogisticsModel.CrossborderTaxQuote) error {
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

func (r *CrossborderTaxQuoteRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.CrossborderTaxQuote, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CrossborderTaxQuote
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type CrossborderTrackingMapRepository struct {
	*Repository[LogisticsModel.CrossborderTrackingMap]
}

func NewCrossborderTrackingMapRepository(db *gorm.DB) *CrossborderTrackingMapRepository {
	return &CrossborderTrackingMapRepository{Repository: NewRepository[LogisticsModel.CrossborderTrackingMap](db)}
}

func (r *CrossborderTrackingMapRepository) Save(ctx context.Context, row *LogisticsModel.CrossborderTrackingMap) error {
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

func (r *CrossborderTrackingMapRepository) List(ctx context.Context, provider string, enabled *bool, limit int) ([]LogisticsModel.CrossborderTrackingMap, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(provider) != "" {
		q = q.Where("provider = ?", strings.TrimSpace(provider))
	}
	if enabled != nil {
		q = q.Where("enabled = ?", *enabled)
	}
	var rows []LogisticsModel.CrossborderTrackingMap
	err = q.Order("priority ASC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CrossborderTrackingMapRepository) GetByUniq(ctx context.Context, provider, providerStatus string) (*LogisticsModel.CrossborderTrackingMap, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CrossborderTrackingMap
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND provider = ? AND provider_status = ?", tenantUUID, strings.TrimSpace(provider), strings.TrimSpace(providerStatus)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
