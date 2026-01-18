package product_sku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pricingmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productskulogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// Service groups the SKU level orchestration entrypoints. Concrete logic will be
// implemented in later phases but we expose strongly typed repositories now so
// downstream handlers can start wiring requests.
type Service struct {
	deps             *app.Deps
	SKURepo          *repo.SKURepository
	AttributeRepo    *repo.AttributeRepository
	ChannelRepo      *repo.ChannelRepository
	InventoryRepo    *repo.InventoryRepository
	AuditLogRepo     *repo.AuditLogRepository
	MediaRepo        *repo.MediaRepository
	BulkTaskRepo     *repo.BulkTaskRepository
	BulkTaskItemRepo *repo.BulkTaskItemRepository
	SerialRepo       *repo.SerialRepository
	logger           *productskulogger.Logger
}

// NewService builds the SKU service scaffold using shared dependencies.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	var skuLogger *productskulogger.Logger
	if entry := deps.RuntimeLogger(deps.Ctx, "product_sku", nil); entry != nil {
		skuLogger = productskulogger.NewLogger(entry)
	}
	return &Service{
		deps:             deps,
		SKURepo:          repo.NewSKURepository(deps.DB),
		AttributeRepo:    repo.NewAttributeRepository(deps.DB),
		ChannelRepo:      repo.NewChannelRepository(deps.DB),
		InventoryRepo:    repo.NewInventoryRepository(deps.DB),
		AuditLogRepo:     repo.NewAuditLogRepository(deps.DB),
		MediaRepo:        repo.NewMediaRepository(deps.DB),
		BulkTaskRepo:     repo.NewBulkTaskRepository(deps.DB),
		BulkTaskItemRepo: repo.NewBulkTaskItemRepository(deps.DB),
		SerialRepo:       repo.NewSerialRepository(deps.DB),
		logger:           skuLogger,
	}
}

// Ready reports whether the service can start serving requests.
func (s *Service) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil
}

// NotImplemented is a helper for unfinished orchestration paths.
func (s *Service) NotImplemented() error {
	return errors.New("product SKU service not implemented yet")
}

// HealthProbe is invoked by router wiring to keep future health checks simple.
func (s *Service) HealthProbe(ctx context.Context) error {
	if !s.Ready() {
		return errors.New("product SKU service not ready")
	}
	return ctx.Err()
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("request context missing")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", errors.New("tenant context missing")
}

func (s *Service) emitEvent(ctx context.Context, action string, payload map[string]any) {
	if s == nil || s.logger == nil {
		return
	}
	var metadata map[string]any
	if len(payload) > 0 {
		metadata = payload
	}
	event := productskulogger.Event{
		Action:   action,
		Metadata: metadata,
	}
	if ctx != nil {
		if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
			event.TenantID = tid
			if payload == nil {
				payload = map[string]any{}
			}
			if _, exists := payload["tenant_id"]; !exists {
				payload["tenant_id"] = tid
			}
			event.Metadata = payload
		}
		if reqID, ok := ctx.Value("request_id").(string); ok && strings.TrimSpace(reqID) != "" {
			event.RequestID = strings.TrimSpace(reqID)
		}
	}
	s.logger.EmitEvent(event)
}

// ListSkus returns paginated SKU summaries for the tenant in context.
func (s *Service) ListSkus(ctx context.Context, query SkuListQuery) (*SkuListResult, error) {
	if s == nil || !s.Ready() {
		return nil, errors.New("product SKU service not ready")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(query.SPUID) != "" {
		hasSpecs, err := s.spuHasSpecGroups(ctx, tenantID, query.SPUID)
		if err != nil {
			return nil, err
		}
		if !hasSpecs {
			page := query.Page
			pageSize := query.PageSize
			if page <= 0 {
				page = 1
			}
			if pageSize <= 0 {
				pageSize = 20
			}
			return &SkuListResult{
				Items:    []SkuListItem{},
				Page:     page,
				PageSize: pageSize,
				Total:    0,
			}, nil
		}
	}
	locale := strings.TrimSpace(query.Locale)
	if locale == "" {
		locale = "zh-CN"
	}
	filters := repo.SkuListFilters{
		SPUID:    strings.TrimSpace(query.SPUID),
		Status:   strings.TrimSpace(query.Status),
		Keyword:  strings.TrimSpace(query.Keyword),
		Locale:   locale,
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	rows, total, err := s.SKURepo.List(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}

	locale = strings.TrimSpace(locale)
	spuNameMap, err := s.resolveSPUNames(ctx, tenantID, rows, locale)
	if err != nil {
		return nil, err
	}

	skuPrices, pbCurrency, err := s.resolveSKUPrices(ctx, tenantID, rows)
	if err != nil {
		return nil, err
	}

	items := make([]SkuListItem, len(rows))
	for i, row := range rows {
		createdAt := row.CreatedAt
		updatedAt := row.UpdatedAt
		specs := parseSkuSpecs(row.SpecValues)
		price := skuPrices[row.ID]
		currency := ""
		if pbCurrency != "" && price != nil {
			currency = pbCurrency
		} else if c, ok := extractCurrencyFromDefaultValues(row.DefaultValues); ok {
			currency = c
		}
		items[i] = SkuListItem{
			ID:          row.ID,
			SPUID:       row.SPUID,
			SPUName:     spuNameMap[row.SPUID],
			SKUCode:     row.SKUCode,
			Status:      row.Status,
			Barcode:     row.Barcode,
			Specs:       specs,
			SpecDisplay: formatSkuSpecsDisplay(specs),
			SalePrice:   price,
			Currency:    currency,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		}
	}
	return &SkuListResult{
		Items:    items,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    total,
	}, nil
}

// GetSku returns a single SKU summary by id.
func (s *Service) GetSku(ctx context.Context, skuID string, locale string) (*SkuListItem, error) {
	if s == nil || !s.Ready() {
		return nil, errors.New("product SKU service not ready")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = "zh-CN"
	}

	row, err := s.SKURepo.FindByID(ctx, tenantID, skuID)
	if err != nil {
		return nil, err
	}
	rows := []productskumodel.ProductSKU{*row}

	spuNameMap, err := s.resolveSPUNames(ctx, tenantID, rows, locale)
	if err != nil {
		return nil, err
	}
	skuPrices, pbCurrency, err := s.resolveSKUPrices(ctx, tenantID, rows)
	if err != nil {
		return nil, err
	}
	specs := parseSkuSpecs(row.SpecValues)
	price := skuPrices[row.ID]
	currency := ""
	if pbCurrency != "" && price != nil {
		currency = pbCurrency
	} else if c, ok := extractCurrencyFromDefaultValues(row.DefaultValues); ok {
		currency = c
	}
	createdAt := row.CreatedAt
	updatedAt := row.UpdatedAt
	item := &SkuListItem{
		ID:          row.ID,
		SPUID:       row.SPUID,
		SPUName:     spuNameMap[row.SPUID],
		SKUCode:     row.SKUCode,
		Status:      row.Status,
		Barcode:     row.Barcode,
		Specs:       specs,
		SpecDisplay: formatSkuSpecsDisplay(specs),
		SalePrice:   price,
		Currency:    currency,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}
	return item, nil
}

type spuNameRow struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type spuLocaleTitleRow struct {
	SPUID string `gorm:"column:spu_id"`
	Title string `gorm:"column:title"`
}

func (s *Service) resolveSPUNames(ctx context.Context, tenantID string, skus []productskumodel.ProductSKU, locale string) (map[string]string, error) {
	result := map[string]string{}
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return result, nil
	}
	seen := map[string]struct{}{}
	for _, sku := range skus {
		id := strings.TrimSpace(sku.SPUID)
		if id == "" {
			continue
		}
		seen[id] = struct{}{}
	}
	if len(seen) == 0 {
		return result, nil
	}
	spuIDs := make([]string, 0, len(seen))
	for id := range seen {
		spuIDs = append(spuIDs, id)
	}

	var localeRows []spuLocaleTitleRow
	if err := s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPULocale{}).
		Select("spu_id, title").
		Where("tenant_uuid = ? AND locale = ? AND spu_id IN ?", tenantID, locale, spuIDs).
		Find(&localeRows).Error; err != nil {
		return nil, err
	}
	for _, row := range localeRows {
		title := strings.TrimSpace(row.Title)
		if title == "" {
			continue
		}
		result[row.SPUID] = title
	}

	missing := make([]string, 0, len(spuIDs))
	for _, id := range spuIDs {
		if _, ok := result[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return result, nil
	}

	var baseRows []spuNameRow
	if err := s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPU{}).
		Select("id, name").
		Where("tenant_uuid = ? AND id IN ?", tenantID, missing).
		Find(&baseRows).Error; err != nil {
		return nil, err
	}
	for _, row := range baseRows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if _, exists := result[row.ID]; !exists {
			result[row.ID] = name
		}
	}
	return result, nil
}

func parseSkuSpecs(raw []byte) []SkuSpec {
	if len(raw) == 0 {
		return nil
	}
	var specs []SkuSpec
	if err := json.Unmarshal(raw, &specs); err != nil {
		return nil
	}
	return specs
}

func formatSkuSpecsDisplay(specs []SkuSpec) string {
	if len(specs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(specs))
	for _, spec := range specs {
		name := strings.TrimSpace(spec.SpecName)
		if name == "" {
			name = strings.TrimSpace(spec.SpecID)
		}
		value := strings.TrimSpace(spec.ValueName)
		if value == "" {
			value = strings.TrimSpace(spec.ValueID)
		}
		switch {
		case name != "" && value != "":
			parts = append(parts, name+": "+value)
		case value != "":
			parts = append(parts, value)
		case name != "":
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, " / ")
}

func (s *Service) resolveSKUPrices(ctx context.Context, tenantID string, skus []productskumodel.ProductSKU) (map[string]*float64, string, error) {
	out := make(map[string]*float64, len(skus))
	if s == nil || s.deps == nil || s.deps.DB == nil || strings.TrimSpace(tenantID) == "" || len(skus) == 0 {
		return out, "", nil
	}
	skuIDs := make([]string, 0, len(skus))
	for _, sku := range skus {
		id := strings.TrimSpace(sku.ID)
		if id == "" {
			continue
		}
		skuIDs = append(skuIDs, id)
	}
	pricebookPrices, pbCurrency, err := s.loadBasePricebookSkuPrices(ctx, tenantID, skuIDs)
	if err != nil {
		return nil, "", err
	}
	for _, sku := range skus {
		if p, ok := pricebookPrices[strings.TrimSpace(sku.ID)]; ok && p != nil {
			out[sku.ID] = p
			continue
		}
		if p, ok := extractSalePriceFromDefaultValues(sku.DefaultValues); ok {
			cp := p
			out[sku.ID] = &cp
		}
	}
	return out, pbCurrency, nil
}

func (s *Service) loadBasePricebookSkuPrices(ctx context.Context, tenantUUID string, skuIDs []string) (map[string]*float64, string, error) {
	out := make(map[string]*float64, len(skuIDs))
	if s == nil || s.deps == nil || s.deps.DB == nil || len(skuIDs) == 0 || strings.TrimSpace(tenantUUID) == "" {
		return out, "", nil
	}
	tx := s.deps.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return out, "", tx.Error
	}
	defer func() { _ = tx.Rollback() }()
	if tx.Dialector != nil && tx.Dialector.Name() != "sqlite" {
		if err := tx.Exec("SELECT set_config('app.tenant_uuid', ?, true)", strings.TrimSpace(tenantUUID)).Error; err != nil {
			return out, "", err
		}
	}

	type pbRow struct {
		ID             string  `gorm:"column:id"`
		Currency       string  `gorm:"column:currency"`
		CurrentVersion *string `gorm:"column:current_version_id"`
	}
	var pb pbRow
	if err := tx.Table(pricingmodel.Pricebook{}.TableName()).
		Select("id, currency, current_version_id").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND code = ? AND status = ?", tenantUUID, "base", "active").
		First(&pb).Error; err != nil {
		return out, "", nil
	}
	versionID := ""
	if pb.CurrentVersion != nil {
		versionID = strings.TrimSpace(*pb.CurrentVersion)
	}
	if versionID == "" {
		return out, strings.TrimSpace(pb.Currency), nil
	}

	type itemRow struct {
		SKUID           string `gorm:"column:sku_id"`
		BaseAmountMinor *int64 `gorm:"column:base_amount_minor"`
		SaleAmountMinor *int64 `gorm:"column:sale_amount_minor"`
	}
	var items []itemRow
	if err := tx.Table(pricingmodel.PricebookItem{}.TableName()).
		Select("sku_id, base_amount_minor, sale_amount_minor").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND version_id = ? AND sku_id IN ?", tenantUUID, versionID, skuIDs).
		Find(&items).Error; err != nil {
		return out, strings.TrimSpace(pb.Currency), err
	}
	for _, it := range items {
		minor := (*int64)(nil)
		if it.SaleAmountMinor != nil {
			minor = it.SaleAmountMinor
		} else if it.BaseAmountMinor != nil {
			minor = it.BaseAmountMinor
		}
		if minor == nil {
			continue
		}
		v := float64(*minor) / 100.0
		out[strings.TrimSpace(it.SKUID)] = &v
	}
	return out, strings.TrimSpace(pb.Currency), nil
}

func extractSalePriceFromDefaultValues(raw []byte) (float64, bool) {
	if len(raw) == 0 {
		return 0, false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, false
	}
	v, ok := m["sale_price"]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, n > 0
	case int:
		return float64(n), n > 0
	case int64:
		return float64(n), n > 0
	case json.Number:
		f, err := n.Float64()
		return f, err == nil && f > 0
	case string:
		f, err := json.Number(strings.TrimSpace(n)).Float64()
		return f, err == nil && f > 0
	default:
		return 0, false
	}
}

func extractCurrencyFromDefaultValues(raw []byte) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", false
	}
	v, ok := m["currency"]
	if !ok || v == nil {
		return "", false
	}
	switch s := v.(type) {
	case string:
		out := strings.TrimSpace(s)
		return out, out != ""
	default:
		out := strings.TrimSpace(fmt.Sprint(v))
		return out, out != ""
	}
}
