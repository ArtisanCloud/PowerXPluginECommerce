package spu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
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
	skuService  *productskuservice.Service
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
		skuService:  productskuservice.NewService(deps),
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
	spuTable := productmodel.SPU{}.TableName()
	versionTable := productmodel.SPUVersion{}.TableName()
	err = s.versionRepo.DB.WithContext(ctx).
		Table(spuTable+" AS spus").
		Select("versions.payload").
		Joins("JOIN "+versionTable+" AS versions ON spus.current_version_id = versions.id").
		Where("spus.tenant_uuid = ? AND spus.id = ?", tenantID, spuID).
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
	if err := s.syncProductSkus(ctx, tenantID, spuID, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SKULinkService) syncProductSkus(ctx context.Context, tenantID, spuID string, linked []LinkedSKU) error {
	if s.skuService == nil || !s.skuService.Ready() {
		return nil
	}
	req := productskuservice.SkuUpsertRequest{
		SKUs: make([]productskuservice.SkuUpsertPayload, 0, len(linked)),
	}
	keepCodes := make(map[string]struct{}, len(linked))
	for _, item := range linked {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			continue
		}
		keepCodes[strings.ToLower(code)] = struct{}{}
		payload := s.composeSkuPayloadFromLink(ctx, tenantID, spuID, item)
		req.SKUs = append(req.SKUs, payload)
	}
	if len(req.SKUs) > 0 {
		if _, err := s.skuService.UpsertSkus(ctx, req); err != nil {
			return err
		}
	}
	return s.pruneProductSkus(ctx, tenantID, spuID, keepCodes)
}

func (s *SKULinkService) composeSkuPayloadFromLink(ctx context.Context, tenantID, spuID string, item LinkedSKU) productskuservice.SkuUpsertPayload {
	specs, defaults, minOrderQty, barcode := parseLinkedSKUMetadata(item)
	status := resolveSKUStatusForSPU(ctx, s, tenantID, spuID)
	payload := productskuservice.SkuUpsertPayload{
		SPUID:         spuID,
		SKUCode:       strings.TrimSpace(item.Code),
		Barcode:       barcode,
		Status:        status,
		MinOrderQty:   minOrderQty,
		Specs:         specs,
		DefaultValues: defaults,
	}
	if payload.MinOrderQty == 0 && payload.DefaultValues.MinOrderQty > 0 {
		payload.MinOrderQty = payload.DefaultValues.MinOrderQty
	}
	if payload.DefaultValues.MinOrderQty == 0 && payload.MinOrderQty > 0 {
		payload.DefaultValues.MinOrderQty = payload.MinOrderQty
	}
	return payload
}

func resolveSKUStatusForSPU(ctx context.Context, s *SKULinkService, tenantID, spuID string) string {
	tenantID = strings.TrimSpace(tenantID)
	spuID = strings.TrimSpace(spuID)
	if tenantID == "" || spuID == "" || s == nil || s.deps == nil || s.deps.DB == nil {
		return "draft"
	}
	var status string
	_ = s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPU{}).
		Select("status").
		Where("tenant_uuid = ? AND id = ?", tenantID, spuID).
		Scan(&status).Error
	if strings.EqualFold(strings.TrimSpace(status), "published") {
		return "online"
	}
	return "draft"
}

func (s *SKULinkService) pruneProductSkus(ctx context.Context, tenantID, spuID string, keep map[string]struct{}) error {
	if s.skuService == nil || s.skuService.SKURepo == nil || s.skuService.SKURepo.DB == nil {
		return nil
	}
	db := s.skuService.SKURepo.DB.WithContext(ctx)
	var skuIDs []string
	query := db.Model(&productskumodel.ProductSKU{}).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID)
	if len(keep) > 0 {
		codes := make([]string, 0, len(keep))
		for code := range keep {
			codes = append(codes, code)
		}
		query = query.Where("LOWER(sku_code) NOT IN ?", codes)
	}
	if err := query.Pluck("id", &skuIDs).Error; err != nil {
		return err
	}
	if len(skuIDs) > 0 && s.skuService.AttributeRepo != nil && s.skuService.AttributeRepo.DB != nil {
		if err := s.skuService.AttributeRepo.DB.WithContext(ctx).
			Where("tenant_uuid = ? AND sku_id IN ?", tenantID, skuIDs).
			Delete(&productskumodel.ProductSKUAttribute{}).Error; err != nil {
			return err
		}
	}
	if len(skuIDs) == 0 {
		return nil
	}
	return db.Where("tenant_uuid = ? AND id IN ?", tenantID, skuIDs).
		Delete(&productskumodel.ProductSKU{}).Error
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
			id = utils.NewUUID()
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

func parseLinkedSKUMetadata(item LinkedSKU) ([]productskuservice.SkuSpec, productskuservice.SkuGeneratorDefaults, int, string) {
	attrs := item.Attributes
	defaults := extractSkuDefaults(attrs)
	if defaults.CostPrice == 0 && item.Pricing.Price > 0 {
		defaults.CostPrice = item.Pricing.Price
	}
	specs := extractSkuSpecs(attrs)
	minOrder := firstInt(attrs, "min_order_qty", "minOrderQty")
	if minOrder == 0 {
		if defaults.MinOrderQty > 0 {
			minOrder = defaults.MinOrderQty
		} else {
			minOrder = firstInt(extractMap(attrs, "defaults", "default_values", "defaultValues"), "min_order_qty", "minOrderQty")
		}
	}
	barcode := firstString(attrs, "barcode", "sku_barcode")
	return specs, defaults, minOrder, barcode
}

func extractSkuSpecs(attrs map[string]any) []productskuservice.SkuSpec {
	if len(attrs) == 0 {
		return nil
	}
	if raw := extractSlice(attrs, "specs", "spec_values", "specValues"); len(raw) > 0 {
		return normalizeSpecArray(raw)
	}
	return fallbackSpecsFromAttributes(attrs)
}

func extractSkuDefaults(attrs map[string]any) productskuservice.SkuGeneratorDefaults {
	var defaults productskuservice.SkuGeneratorDefaults
	if len(attrs) == 0 {
		return defaults
	}
	mergeDefaults(&defaults, attrs)
	if nested := extractMap(attrs, "defaults", "default_values", "defaultValues"); len(nested) > 0 {
		mergeDefaults(&defaults, nested)
	}
	if logistics := extractMap(attrs, "logistics"); len(logistics) > 0 {
		mergeDefaults(&defaults, logistics)
	}
	return defaults
}

func mergeDefaults(dst *productskuservice.SkuGeneratorDefaults, src map[string]any) {
	if dst == nil || len(src) == 0 {
		return
	}
	if prefix := firstString(src, "barcode_prefix", "barcodePrefix"); prefix != "" {
		dst.BarcodePrefix = prefix
	}
	if qty := firstInt(src, "min_order_qty", "minOrderQty"); qty > 0 {
		dst.MinOrderQty = qty
	}
	if cost := firstFloat(src, "cost_price", "costPrice", "price"); cost > 0 {
		dst.CostPrice = cost
	}
	if weight := firstFloat(src, "weight"); weight > 0 {
		dst.Weight = weight
	}
	if dims := firstString(src, "dimensions", "dimension", "size"); dims != "" {
		dst.Dimensions = dims
	}
}

func extractSlice(attrs map[string]any, keys ...string) []any {
	for _, key := range keys {
		if raw, ok := attrs[key]; ok {
			switch val := raw.(type) {
			case []any:
				return val
			case []map[string]any:
				res := make([]any, len(val))
				for i := range val {
					res[i] = val[i]
				}
				return res
			case []productskuservice.SkuSpec:
				res := make([]any, len(val))
				for i := range val {
					res[i] = val[i]
				}
				return res
			}
		}
	}
	return nil
}

func normalizeSpecArray(items []any) []productskuservice.SkuSpec {
	result := make([]productskuservice.SkuSpec, 0, len(items))
	for idx, raw := range items {
		var spec productskuservice.SkuSpec
		switch val := raw.(type) {
		case productskuservice.SkuSpec:
			spec = val
		case map[string]any:
			spec = mapToSpec(val, idx)
		case map[string]string:
			converted := make(map[string]any, len(val))
			for k, v := range val {
				converted[k] = v
			}
			spec = mapToSpec(converted, idx)
		default:
			continue
		}
		if spec.SpecID == "" && spec.SpecName != "" {
			spec.SpecID = slugify(spec.SpecName, fmt.Sprintf("spec-%d", idx+1))
		}
		if spec.ValueID == "" && spec.ValueName != "" {
			spec.ValueID = slugify(spec.ValueName, fmt.Sprintf("%s-%d", spec.SpecID, idx+1))
		}
		if spec.SpecID == "" || spec.ValueID == "" {
			continue
		}
		if spec.SpecName == "" {
			spec.SpecName = spec.SpecID
		}
		if spec.ValueName == "" {
			spec.ValueName = spec.ValueID
		}
		result = append(result, spec)
	}
	return result
}

var reservedAttributeKeys = map[string]struct{}{
	"defaults":       {},
	"default_values": {},
	"defaultValues":  {},
	"pricing":        {},
	"inventoryRef":   {},
	"inventory_ref":  {},
	"cloneFrom":      {},
	"clone_from":     {},
	"min_order_qty":  {},
	"minOrderQty":    {},
	"logistics":      {},
	"weight":         {},
	"dimensions":     {},
	"dimension":      {},
	"size":           {},
	"barcode":        {},
	"sku_barcode":    {},
}

func fallbackSpecsFromAttributes(attrs map[string]any) []productskuservice.SkuSpec {
	if len(attrs) == 0 {
		return nil
	}
	idx := 0
	result := make([]productskuservice.SkuSpec, 0, len(attrs))
	for key, raw := range attrs {
		if _, reserved := reservedAttributeKeys[key]; reserved {
			continue
		}
		value := stringValue(raw)
		if value == "" {
			continue
		}
		idx++
		specID := slugify(key, fmt.Sprintf("attr-%d", idx))
		valueID := slugify(value, fmt.Sprintf("%s-%d", specID, idx))
		result = append(result, productskuservice.SkuSpec{
			SpecID:    specID,
			SpecName:  key,
			ValueID:   valueID,
			ValueName: value,
		})
	}
	return result
}

func mapToSpec(src map[string]any, idx int) productskuservice.SkuSpec {
	spec := productskuservice.SkuSpec{}
	spec.SpecID = firstString(src, "spec_id", "specId", "id", "code", "spec")
	spec.SpecName = firstString(src, "spec_name", "specName", "name", "label", "title")
	spec.ValueID = firstString(src, "value_id", "valueId", "value_code", "valueCode", "value")
	spec.ValueName = firstString(src, "value_name", "valueName", "label", "name")
	if spec.ValueName == "" {
		spec.ValueName = spec.ValueID
	}
	if spec.ValueID == "" && spec.ValueName != "" {
		spec.ValueID = slugify(spec.ValueName, fmt.Sprintf("val-%d", idx+1))
	}
	return spec
}

func extractMap(attrs map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if raw, ok := attrs[key]; ok {
			switch val := raw.(type) {
			case map[string]any:
				return val
			case map[string]string:
				converted := make(map[string]any, len(val))
				for k, v := range val {
					converted[k] = v
				}
				return converted
			}
		}
	}
	return nil
}

func firstString(attrs map[string]any, keys ...string) string {
	if len(attrs) == 0 {
		return ""
	}
	for _, key := range keys {
		if val, ok := attrs[key]; ok {
			if str := stringValue(val); str != "" {
				return str
			}
		}
	}
	return ""
}

func firstInt(attrs map[string]any, keys ...string) int {
	if len(attrs) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, ok := attrs[key]; ok {
			if parsed, ok := intValue(val); ok {
				return parsed
			}
		}
	}
	return 0
}

func firstFloat(attrs map[string]any, keys ...string) float64 {
	if len(attrs) == 0 {
		return 0
	}
	for _, key := range keys {
		if val, ok := attrs[key]; ok {
			if parsed, ok := floatValue(val); ok {
				return parsed
			}
		}
	}
	return 0
}

func stringValue(val any) string {
	switch typed := val.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return strings.TrimSpace(typed.String())
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func intValue(val any) (int, bool) {
	switch typed := val.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case int32:
		return int(typed), true
	case uint:
		return int(typed), true
	case uint64:
		return int(typed), true
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	case json.Number:
		if iv, err := strconv.Atoi(typed.String()); err == nil {
			return iv, true
		}
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		if strings.Contains(trimmed, ".") {
			if fv, err := strconv.ParseFloat(trimmed, 64); err == nil {
				return int(fv), true
			}
		} else if iv, err := strconv.Atoi(trimmed); err == nil {
			return iv, true
		}
	}
	return 0, false
}

func floatValue(val any) (float64, bool) {
	switch typed := val.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		if fv, err := typed.Float64(); err == nil {
			return fv, true
		}
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		if fv, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return fv, true
		}
	}
	return 0, false
}

func slugify(input, fallback string) string {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if trimmed == "" {
		return fallback
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return fallback
	}
	return result
}
