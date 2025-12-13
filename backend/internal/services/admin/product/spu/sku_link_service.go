package spu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/domain/repository/product"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LinkedSKU represents a simplified SKU record persisted inside the SPU version payload.
type LinkedSKU struct {
	ID           string         `json:"id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	InventoryRef string         `json:"inventoryRef,omitempty"`
	Pricing      SKUPricing     `json:"pricing"`
	SourceSKU    string         `json:"sourceSku,omitempty"`
}

// SKUPricing captures a snapshot of pricing metadata inherited from the pricing service.
type SKUPricing struct {
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

// ReplaceSKUsRequest represents a batch payload for SKU linking.
type ReplaceSKUsRequest struct {
	Items []SKULinkInput `json:"items"`
}

// SKULinkInput describes a single SKU entry to be created or cloned.
type SKULinkInput struct {
	ID           string         `json:"id,omitempty"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	InventoryRef string         `json:"inventoryRef,omitempty"`
	Pricing      SKUPricing     `json:"pricing"`
	CloneFrom    string         `json:"cloneFrom,omitempty"`
}

// SKULinkService manages SKU associations scoped to an SPU version payload.
type SKULinkService struct {
	deps        *app.Deps
	spuRepo     *productrepo.SPURepository
	versionRepo *productrepo.VersionRepository
}

// NewSKULinkService constructs the SKU service with shared dependencies.
func NewSKULinkService(deps *app.Deps) *SKULinkService {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &SKULinkService{
		deps:        deps,
		spuRepo:     productrepo.NewSPURepository(deps.DB),
		versionRepo: productrepo.NewVersionRepository(deps.DB),
	}
}

// List returns SKU links stored in the current version payload.
func (s *SKULinkService) List(ctx context.Context, spuID string) ([]LinkedSKU, error) {
	if s == nil {
		return nil, errors.New("sku link service unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var versionPayload struct {
		Payload []byte
	}
	err = s.versionRepo.DB.WithContext(ctx).
		Table("product_spus").
		Select("product_spu_versions.payload").
		Joins("JOIN product_spu_versions ON product_spus.current_version_id = product_spu_versions.id").
		Where("product_spus.tenant_uuid = ? AND product_spus.id = ?", tenantID, spuID).
		Limit(1).
		Take(&versionPayload).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []LinkedSKU{}, nil
		}
		return nil, err
	}
	decoded := decodeVersionPayload(datatypes.JSON(versionPayload.Payload))
	return parseLinkedSKUs(decoded), nil
}

// Replace overwrites SKU associations for the given SPU.
func (s *SKULinkService) Replace(ctx context.Context, spuID string, req ReplaceSKUsRequest) ([]LinkedSKU, error) {
	if s == nil {
		return nil, errors.New("sku link service unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range req.Items {
		if strings.TrimSpace(item.Code) == "" {
			return nil, fmt.Errorf("sku code is required")
		}
		if strings.TrimSpace(item.Name) == "" {
			return nil, fmt.Errorf("sku name is required for %s", item.Code)
		}
	}
	var result []LinkedSKU
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
			return err
		}
		versionID := derefString(spu.CurrentVersionID)
		if versionID == "" {
			return errors.New("spu missing version payload")
		}
		var version productmodel.SPUVersion
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, versionID).First(&version).Error; err != nil {
			return err
		}
		payload := decodeVersionPayload(version.Payload)
		existing := parseLinkedSKUs(payload)
		linked, err := s.buildLinkedSKUs(req.Items, existing)
		if err != nil {
			return err
		}
		payload["skus"] = linked
		updatedPayload := encodeVersionPayload(payload)
		if err := tx.Model(&productmodel.SPUVersion{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, versionID).
			Update("payload", updatedPayload).Error; err != nil {
			return err
		}
		result = linked
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SKULinkService) buildLinkedSKUs(inputs []SKULinkInput, existing []LinkedSKU) ([]LinkedSKU, error) {
	if len(inputs) == 0 {
		return existing, nil
	}
	existingMap := make(map[string]LinkedSKU)
	for _, item := range existing {
		if item.ID != "" {
			existingMap[item.ID] = item
		}
		if item.Code != "" {
			existingMap[item.Code] = item
		}
	}
	codes := map[string]struct{}{}
	linked := make([]LinkedSKU, 0, len(inputs))
	for _, input := range inputs {
		code := strings.TrimSpace(input.Code)
		if code == "" {
			return nil, fmt.Errorf("sku code is required")
		}
		lower := strings.ToLower(code)
		if _, ok := codes[lower]; ok {
			return nil, fmt.Errorf("duplicate sku code: %s", code)
		}
		codes[lower] = struct{}{}
		template := LinkedSKU{}
		if input.CloneFrom != "" {
			if val, ok := existingMap[input.CloneFrom]; ok {
				template = val
			} else {
				return nil, fmt.Errorf("clone source %s not found", input.CloneFrom)
			}
		}
		id := input.ID
		if strings.TrimSpace(id) == "" {
			id = uuid.NewString()
		}
		pricing := input.Pricing
		if pricing.Currency == "" {
			if template.Pricing.Currency != "" {
				pricing.Currency = template.Pricing.Currency
			} else {
				pricing.Currency = "CNY"
			}
		}
		if pricing.Price == 0 && template.Pricing.Price > 0 {
			pricing.Price = template.Pricing.Price
		}
		attrs := input.Attributes
		if len(attrs) == 0 && len(template.Attributes) > 0 {
			attrs = template.Attributes
		}
		linked = append(linked, LinkedSKU{
			ID:           id,
			Code:         code,
			Name:         firstNonEmpty(input.Name, template.Name),
			Attributes:   attrs,
			InventoryRef: firstNonEmpty(input.InventoryRef, template.InventoryRef),
			Pricing:      pricing,
			SourceSKU:    input.CloneFrom,
		})
	}
	return linked, nil
}

func (s *SKULinkService) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = s.deps.Ctx
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && tid != "" {
		return tid, nil
	}
	return "", ErrMissingTenant
}

func parseLinkedSKUs(payload map[string]any) []LinkedSKU {
	raw, ok := payload["skus"]
	if !ok || raw == nil {
		return []LinkedSKU{}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return []LinkedSKU{}
	}
	var items []LinkedSKU
	if err := json.Unmarshal(data, &items); err != nil {
		return []LinkedSKU{}
	}
	return items
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
