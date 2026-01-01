package product_sku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var allowedStatuses = map[string]struct{}{
	"draft":   {},
	"ready":   {},
	"online":  {},
	"offline": {},
}

// UpsertSkus creates SKU records from generator selections.
func (s *Service) UpsertSkus(ctx context.Context, req SkuUpsertRequest) (*SkuUpsertResult, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	if len(req.SKUs) == 0 {
		return nil, errors.New("payload skus is empty")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existingMap := map[string]productskumodel.ProductSKU{}
	existingCodes := map[string]struct{}{}
	existingRecords := map[string]productskumodel.ProductSKU{}

	// Preload existing combos grouped by SPU to prevent duplicates.
	spuGroups := make(map[string]struct{})
	for _, sku := range req.SKUs {
		if strings.TrimSpace(sku.SPUID) == "" {
			return nil, errors.New("spu_id is required for each sku")
		}
		spuGroups[sku.SPUID] = struct{}{}
	}
	for spuID := range spuGroups {
		combos, codes, records, err := s.lookupExistingSpecHashes(ctx, tenantID, spuID)
		if err != nil {
			return nil, err
		}
		for hash, rec := range combos {
			existingMap[keyWithGroup(spuID, hash)] = rec
		}
		for code := range codes {
			existingCodes[keyWithGroup(spuID, code)] = struct{}{}
		}
		for code, rec := range records {
			existingRecords[keyWithGroup(spuID, code)] = rec
		}
	}

	result := &SkuUpsertResult{
		Skipped:   []string{},
		Summaries: []SkuSummary{},
	}

	err = s.SKURepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for _, payload := range req.SKUs {
			spuID := strings.TrimSpace(payload.SPUID)
			if spuID == "" {
				return errors.New("spu_id is required for each sku")
			}
			code := strings.TrimSpace(payload.SKUCode)
			if code == "" {
				return fmt.Errorf("sku_code is required for sku under spu %s", spuID)
			}
			lowerCode := strings.ToLower(code)
			codeKey := keyWithGroup(spuID, lowerCode)
			specHash := hashSpecValues(payload.Specs)
			hasSpecs := len(payload.Specs) > 0
			if hasSpecs && specHash != "" {
				if existing, ok := existingMap[keyWithGroup(spuID, specHash)]; ok && !strings.EqualFold(existing.SKUCode, code) {
					result.Skipped = append(result.Skipped, payload.SKUCode)
					continue
				}
			}
			if record, ok := existingRecords[codeKey]; ok {
				if err := s.updateExistingSKU(tx, tenantID, record, payload, now); err != nil {
					return err
				}
				removeSpecHashForSKU(existingMap, spuID, record.SKUCode)
				if hasSpecs && specHash != "" {
					record.SKUCode = code
					record.SPUID = spuID
					existingMap[keyWithGroup(spuID, specHash)] = record
				}
				result.Summaries = append(result.Summaries, SkuSummary{
					ID:      record.ID,
					SPUID:   record.SPUID,
					SKUCode: code,
					Status:  sanitizeStatus(payload.Status),
					Specs:   payload.Specs,
				})
				continue
			}
			if _, conflict := existingCodes[codeKey]; conflict {
				result.Skipped = append(result.Skipped, payload.SKUCode)
				continue
			}
			record := &productskumodel.ProductSKU{
				ID:             utils.NewUUID(),
				TenantUUID:     tenantID,
				SPUID:          spuID,
				SKUCode:        code,
				Barcode:        payload.Barcode,
				Status:         sanitizeStatus(payload.Status),
				MinOrderQty:    payload.MinOrderQty,
				SpecValues:     marshalJSON(payload.Specs),
				DefaultValues:  marshalJSON(payload.DefaultValues),
				CreatedAt:      now,
				UpdatedAt:      now,
				LifecyclePhase: "concept",
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
			if err := s.replaceSkuAttributes(tx, tenantID, record.ID, payload.Specs); err != nil {
				return err
			}
			if hasSpecs && specHash != "" {
				existingMap[keyWithGroup(spuID, specHash)] = *record
			}
			existingCodes[codeKey] = struct{}{}
			existingRecords[codeKey] = *record
			result.Created++
			result.Summaries = append(result.Summaries, SkuSummary{
				ID:      record.ID,
				SPUID:   record.SPUID,
				SKUCode: record.SKUCode,
				Status:  record.Status,
				Specs:   payload.Specs,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) updateExistingSKU(tx *gorm.DB, tenantID string, existing productskumodel.ProductSKU, payload SkuUpsertPayload, now time.Time) error {
	updates := map[string]interface{}{
		"sku_code":       payload.SKUCode,
		"barcode":        payload.Barcode,
		"status":         sanitizeStatus(payload.Status),
		"min_order_qty":  payload.MinOrderQty,
		"spec_values":    marshalJSON(payload.Specs),
		"default_values": marshalJSON(payload.DefaultValues),
		"updated_at":     now,
	}
	if err := tx.Model(&productskumodel.ProductSKU{}).
		Where("id = ?", existing.ID).
		Updates(updates).Error; err != nil {
		return err
	}
	return s.replaceSkuAttributes(tx, tenantID, existing.ID, payload.Specs)
}

func (s *Service) replaceSkuAttributes(tx *gorm.DB, tenantID, skuID string, specs []SkuSpec) error {
	if err := tx.Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID).
		Delete(&productskumodel.ProductSKUAttribute{}).Error; err != nil {
		return err
	}
	for idx, spec := range specs {
		if strings.TrimSpace(spec.SpecID) == "" || strings.TrimSpace(spec.ValueID) == "" {
			continue
		}
		attr := &productskumodel.ProductSKUAttribute{
			ID:           utils.NewUUID(),
			TenantUUID:   tenantID,
			SKUId:        skuID,
			SpecID:       spec.SpecID,
			SpecValueID:  spec.ValueID,
			SpecName:     spec.SpecName,
			ValueName:    spec.ValueName,
			DisplayOrder: idx,
		}
		if err := tx.Create(attr).Error; err != nil {
			return err
		}
	}
	return nil
}

func removeSpecHashForSKU(entries map[string]productskumodel.ProductSKU, spuID, code string) {
	if code == "" {
		return
	}
	for key, rec := range entries {
		if rec.SPUID == spuID && strings.EqualFold(rec.SKUCode, code) {
			delete(entries, key)
			break
		}
	}
}

func sanitizeStatus(candidate string) string {
	status := strings.ToLower(strings.TrimSpace(candidate))
	if status == "" {
		return "draft"
	}
	if _, ok := allowedStatuses[status]; ok {
		return status
	}
	return "draft"
}

func marshalJSON(v any) datatypes.JSON {
	if v == nil {
		return datatypes.JSON([]byte("null"))
	}
	body, _ := json.Marshal(v)
	return datatypes.JSON(body)
}

func keyWithGroup(group, hash string) string {
	return group + "::" + hash
}
