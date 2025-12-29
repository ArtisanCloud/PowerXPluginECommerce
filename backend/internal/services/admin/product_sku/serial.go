package product_sku

import (
	"context"
	"errors"
	"strings"
	"time"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

const (
	serialStatusAvailable = "available"
)

// UpsertSerialRecord records serial or batch level metadata for compliance.
func (s *Service) UpsertSerialRecord(ctx context.Context, skuID string, input SerialRecordInput) (*SerialRecordDTO, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	if s.SerialRepo == nil || s.SerialRepo.DB == nil {
		return nil, errors.New("serial repository unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	serial := strings.TrimSpace(input.SerialNo)
	if serial == "" {
		return nil, errors.New("serial_no is required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = serialStatusAvailable
	}

	record := productskumodel.ProductSKUSerialRecord{
		ID:         utils.NewUUID(),
		TenantUUID: tenantID,
		SKUId:      skuID,
		SerialNo:   serial,
		BatchNo:    strings.TrimSpace(input.BatchNo),
		Status:     status,
		ExpiresAt:  input.ExpiresAt,
		CreatedAt:  time.Now(),
	}

	db := s.SerialRepo.DB.WithContext(ctx)
	err = db.Clauses(s.SerialRepo.OnConflictDoNothing()).Create(&record).Error
	if err != nil {
		return nil, err
	}
	// If record already existed, fetch existing to return
	var existing productskumodel.ProductSKUSerialRecord
	if err := db.Where("tenant_uuid = ? AND sku_id = ? AND serial_no = ?", tenantID, skuID, serial).First(&existing).Error; err != nil {
		return nil, err
	}
	updates := map[string]any{
		"batch_no":   strings.TrimSpace(input.BatchNo),
		"status":     status,
		"expires_at": input.ExpiresAt,
	}
	if err := db.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := db.Where("id = ?", existing.ID).First(&existing).Error; err != nil {
		return nil, err
	}
	s.emitEvent(ctx, "serial_upserted", map[string]any{
		"tenant_id": tenantID,
		"sku_id":    skuID,
		"serial_no": serial,
		"status":    status,
	})
	dto := mapSerialRecord(existing)
	return &dto, nil
}

// ListSerialRecords fetches historical serial numbers for a SKU.
func (s *Service) ListSerialRecords(ctx context.Context, skuID string, filters SerialRecordFilters) ([]SerialRecordDTO, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	if s.SerialRepo == nil || s.SerialRepo.DB == nil {
		return nil, errors.New("serial repository unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	query := s.SerialRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID)
	if v := strings.TrimSpace(filters.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := strings.TrimSpace(filters.Batch); v != "" {
		query = query.Where("batch_no = ?", v)
	}
	limit := filters.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query = query.Order("created_at DESC").Limit(limit)
	var records []productskumodel.ProductSKUSerialRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	dtos := make([]SerialRecordDTO, 0, len(records))
	for _, rec := range records {
		dto := mapSerialRecord(rec)
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func mapSerialRecord(record productskumodel.ProductSKUSerialRecord) SerialRecordDTO {
	return SerialRecordDTO{
		ID:        record.ID,
		SKUID:     record.SKUId,
		SerialNo:  record.SerialNo,
		BatchNo:   record.BatchNo,
		Status:    record.Status,
		ExpiresAt: record.ExpiresAt,
		AuditLog:  record.AuditLogID,
		CreatedAt: record.CreatedAt,
	}
}
