package sellability

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	pricingmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var miniAppSellableSKUStatuses = map[string]struct{}{
	"online":    {},
	"ready":     {},
	"published": {},
}

// Service aggregates sellability gates (price/inventory/channel visibility) for mini-app.
type Service struct {
	db    *gorm.DB
	nowFn func() time.Time
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:    db,
		nowFn: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) WithNow(nowFn func() time.Time) *Service {
	if s == nil {
		return nil
	}
	if nowFn != nil {
		s.nowFn = nowFn
	}
	return s
}

func (s *Service) Evaluate(ctx context.Context, tenantUUID, spuID, channel, locale string) (*ResultDTO, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("sellability service unavailable")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	spuID = strings.TrimSpace(spuID)
	channel = strings.TrimSpace(channel)
	_ = strings.TrimSpace(locale)
	if tenantUUID == "" {
		return nil, errors.New("tenant context missing")
	}
	if spuID == "" {
		return nil, errors.New("spu id is required")
	}
	if channel == "" {
		return nil, errors.New("channel is required")
	}

	tx := s.db.WithContext(ctx)
	var spu productmodel.SPU
	if err := tx.Where("tenant_uuid = ? AND deleted_at IS NULL AND id = ?", tenantUUID, spuID).First(&spu).Error; err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(spu.Status), "published") {
		return nil, gorm.ErrRecordNotFound
	}

	skus, err := s.loadSKUs(tx, tenantUUID, spuID)
	if err != nil {
		return nil, err
	}

	channelCfg, err := s.loadChannelVisibility(tx, tenantUUID, spuID, channel)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		channelCfg = nil
	}
	items, err := s.evalSKUItems(tx, tenantUUID, skus, map[string]*productmodel.ChannelVisibility{spuID: channelCfg})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool { return items[i].SKUID < items[j].SKUID })
	return &ResultDTO{
		SPUID:   spuID,
		Channel: channel,
		Items:   items,
	}, nil
}

// EvaluateSKUs evaluates sellability for a list of SKU IDs under the same channel.
// This is designed for order creation flows which only have SKU IDs at submission time.
func (s *Service) EvaluateSKUs(ctx context.Context, tenantUUID string, skuIDs []string, channel, locale string) (*SKUResultDTO, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("sellability service unavailable")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	channel = strings.TrimSpace(channel)
	_ = strings.TrimSpace(locale)
	if tenantUUID == "" {
		return nil, errors.New("tenant context missing")
	}
	if len(skuIDs) == 0 {
		return nil, errors.New("sku ids are required")
	}
	if channel == "" {
		return nil, errors.New("channel is required")
	}

	cleanIDs := make([]string, 0, len(skuIDs))
	seen := map[string]struct{}{}
	for _, raw := range skuIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleanIDs = append(cleanIDs, id)
	}
	if len(cleanIDs) == 0 {
		return nil, errors.New("sku ids are required")
	}

	tx := s.db.WithContext(ctx)
	skus, err := s.loadSKUsByID(tx, tenantUUID, cleanIDs)
	if err != nil {
		return nil, err
	}

	spuIDs := make([]string, 0, len(skus))
	spuSeen := map[string]struct{}{}
	for _, sku := range skus {
		spuID := strings.TrimSpace(sku.SPUID)
		if spuID == "" {
			continue
		}
		if _, ok := spuSeen[spuID]; ok {
			continue
		}
		spuSeen[spuID] = struct{}{}
		spuIDs = append(spuIDs, spuID)
	}

	channelCfgBySPU, err := s.loadChannelVisibilities(tx, tenantUUID, spuIDs, channel)
	if err != nil {
		return nil, err
	}

	publishedSPU := map[string]bool{}
	if len(spuIDs) > 0 {
		type row struct {
			ID     string `gorm:"column:id"`
			Status string `gorm:"column:status"`
		}
		var rows []row
		if err := tx.Table(productmodel.SPU{}.TableName()).
			Select("id, status").
			Where("tenant_uuid = ? AND deleted_at IS NULL AND id IN ?", tenantUUID, spuIDs).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			publishedSPU[strings.TrimSpace(r.ID)] = strings.EqualFold(strings.TrimSpace(r.Status), "published")
		}
	}

	items, err := s.evalSKUItems(tx, tenantUUID, skus, channelCfgBySPU)
	if err != nil {
		return nil, err
	}
	for i := range items {
		skuID := strings.TrimSpace(items[i].SKUID)
		spuID := ""
		for _, sku := range skus {
			if strings.TrimSpace(sku.ID) == skuID {
				spuID = strings.TrimSpace(sku.SPUID)
				break
			}
		}
		if spuID != "" && !publishedSPU[spuID] {
			items[i].Reasons = appendReason(items[i].Reasons, string(ReasonSPUNotPublished))
			items[i].Sellable = false
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].SKUID < items[j].SKUID })
	return &SKUResultDTO{
		Channel: channel,
		Items:   items,
	}, nil
}

func (s *Service) EvaluateSummaries(ctx context.Context, tenantUUID string, spuIDs []string, channel, locale string) (map[string]SummaryDTO, error) {
	out := make(map[string]SummaryDTO, len(spuIDs))
	if len(spuIDs) == 0 {
		return out, nil
	}
	for _, id := range spuIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		result, err := s.Evaluate(ctx, tenantUUID, id, channel, locale)
		if err != nil {
			out[id] = SummaryDTO{Sellable: false, Reasons: []string{string(ReasonUnknown)}}
			continue
		}
		out[id] = summarizeResult(result)
	}
	return out, nil
}

func summarizeResult(result *ResultDTO) SummaryDTO {
	if result == nil || len(result.Items) == 0 {
		return SummaryDTO{Sellable: false, Reasons: []string{string(ReasonUnknown)}}
	}
	for _, it := range result.Items {
		if it.Sellable {
			return SummaryDTO{Sellable: true, Reasons: []string{}}
		}
	}

	precedence := []string{
		string(ReasonChannelDisabled),
		string(ReasonChannelStatusBlocked),
		string(ReasonNotInAvailabilityWindow),
		string(ReasonSKUNotOnline),
		string(ReasonNoPublicPrice),
		string(ReasonOutOfStock),
	}

	present := map[string]struct{}{}
	for _, it := range result.Items {
		for _, r := range it.Reasons {
			present[r] = struct{}{}
		}
	}
	for _, p := range precedence {
		if _, ok := present[p]; ok {
			return SummaryDTO{Sellable: false, Reasons: []string{p}}
		}
	}
	return SummaryDTO{Sellable: false, Reasons: []string{string(ReasonUnknown)}}
}

type skuRow struct {
	ID            string         `gorm:"column:id"`
	SPUID         string         `gorm:"column:spu_id"`
	Status        string         `gorm:"column:status"`
	SKUCode       string         `gorm:"column:sku_code"`
	DefaultValues datatypes.JSON `gorm:"column:default_values"`
}

func (s *Service) loadSKUs(tx *gorm.DB, tenantUUID, spuID string) ([]skuRow, error) {
	var rows []skuRow
	if err := tx.Table(productskumodel.ProductSKU{}.TableName()).
		Select("id, spu_id, status, sku_code, default_values").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND spu_id = ?", tenantUUID, spuID).
		Order("created_at ASC, sku_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) loadSKUsByID(tx *gorm.DB, tenantUUID string, skuIDs []string) ([]skuRow, error) {
	var rows []skuRow
	if len(skuIDs) == 0 {
		return rows, nil
	}
	if err := tx.Table(productskumodel.ProductSKU{}.TableName()).
		Select("id, spu_id, status, sku_code, default_values").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND id IN ?", tenantUUID, skuIDs).
		Order("created_at ASC, sku_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) loadChannelVisibility(tx *gorm.DB, tenantUUID, spuID, channel string) (*productmodel.ChannelVisibility, error) {
	var row productmodel.ChannelVisibility
	err := tx.Table(productmodel.ChannelVisibility{}.TableName()).
		Where("tenant_uuid = ? AND spu_id = ? AND channel = ?", tenantUUID, spuID, channel).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadChannelVisibilities(tx *gorm.DB, tenantUUID string, spuIDs []string, channel string) (map[string]*productmodel.ChannelVisibility, error) {
	out := make(map[string]*productmodel.ChannelVisibility, len(spuIDs))
	if len(spuIDs) == 0 {
		return out, nil
	}
	var rows []productmodel.ChannelVisibility
	if err := tx.Table(productmodel.ChannelVisibility{}.TableName()).
		Where("tenant_uuid = ? AND channel = ? AND spu_id IN ?", tenantUUID, channel, spuIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		r := rows[i]
		out[strings.TrimSpace(r.SPUID)] = &r
	}
	return out, nil
}

func (s *Service) evalChannelGate(cfg *productmodel.ChannelVisibility) []string {
	now := time.Now().UTC()
	if s != nil && s.nowFn != nil {
		now = s.nowFn()
	}

	if cfg == nil {
		return []string{string(ReasonChannelDisabled)}
	}

	availability := strings.ToLower(strings.TrimSpace(cfg.Availability))
	switch availability {
	case "unlisted", "withheld":
		return []string{string(ReasonChannelDisabled)}
	}

	audit := strings.ToLower(strings.TrimSpace(cfg.AuditState))
	if audit != "" && audit != "approved" {
		return []string{string(ReasonChannelStatusBlocked)}
	}

	// Apply window checks when timestamps are provided; if a channel is "scheduled" without
	// a valid window, treat it as unavailable.
	if availability == "scheduled" && cfg.PublishAt == nil && cfg.WithdrawAt == nil {
		return []string{string(ReasonNotInAvailabilityWindow)}
	}
	if cfg.PublishAt != nil && now.Before(*cfg.PublishAt) {
		return []string{string(ReasonNotInAvailabilityWindow)}
	}
	if cfg.WithdrawAt != nil && !now.Before(*cfg.WithdrawAt) {
		return []string{string(ReasonNotInAvailabilityWindow)}
	}
	return []string{}
}

func (s *Service) loadAvailableQty(tx *gorm.DB, tenantUUID string, skuIDs []string) map[string]int {
	out := make(map[string]int, len(skuIDs))
	if len(skuIDs) == 0 {
		return out
	}
	type row struct {
		SKUID string `gorm:"column:sku_id"`
		Qty   int64  `gorm:"column:qty"`
	}
	var rows []row
	if err := tx.Table(productskumodel.ProductSKUInventory{}.TableName()).
		Select("sku_id, SUM(available_qty - locked_qty) AS qty").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND sku_id IN ?", tenantUUID, skuIDs).
		Group("sku_id").
		Scan(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		id := strings.TrimSpace(r.SKUID)
		if id == "" {
			continue
		}
		qty := r.Qty
		if qty < 0 {
			qty = 0
		}
		maxInt := int64(^uint(0) >> 1)
		if qty > maxInt {
			qty = maxInt
		}
		out[id] = int(qty)
	}
	return out
}

func (s *Service) loadBasePricebookSkuPrices(tx *gorm.DB, tenantUUID string, skuIDs []string) (map[string]*float64, string) {
	out := make(map[string]*float64, len(skuIDs))
	if len(skuIDs) == 0 {
		return out, ""
	}
	type pbRow struct {
		ID               string  `gorm:"column:id"`
		Currency         string  `gorm:"column:currency"`
		CurrentVersionID *string `gorm:"column:current_version_id"`
	}
	var pb pbRow
	if err := tx.Table(pricingmodel.Pricebook{}.TableName()).
		Select("id, currency, current_version_id").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND code = ? AND status = ?", tenantUUID, "base", "active").
		First(&pb).Error; err != nil {
		return out, ""
	}
	versionID := ""
	if pb.CurrentVersionID != nil {
		versionID = strings.TrimSpace(*pb.CurrentVersionID)
	}
	if versionID == "" {
		return out, strings.TrimSpace(pb.Currency)
	}

	type itemRow struct {
		SKUID           string `gorm:"column:sku_id"`
		BaseAmountMinor *int64 `gorm:"column:base_amount_minor"`
		SaleAmountMinor *int64 `gorm:"column:sale_amount_minor"`
		MsrpAmountMinor *int64 `gorm:"column:msrp_amount_minor"`
	}
	var items []itemRow
	if err := tx.Table(pricingmodel.PricebookItem{}.TableName()).
		Select("sku_id, base_amount_minor, sale_amount_minor, msrp_amount_minor").
		Where("tenant_uuid = ? AND deleted_at IS NULL AND version_id = ? AND sku_id IN ?", tenantUUID, versionID, skuIDs).
		Find(&items).Error; err != nil {
		return out, strings.TrimSpace(pb.Currency)
	}
	for _, it := range items {
		minor := (*int64)(nil)
		if it.SaleAmountMinor != nil {
			minor = it.SaleAmountMinor
		} else if it.BaseAmountMinor != nil {
			minor = it.BaseAmountMinor
		} else if it.MsrpAmountMinor != nil {
			minor = it.MsrpAmountMinor
		}
		if minor == nil {
			continue
		}
		v := float64(*minor) / 100.0
		out[strings.TrimSpace(it.SKUID)] = &v
	}
	return out, strings.TrimSpace(pb.Currency)
}

func (s *Service) resolveSKUPrice(sku skuRow, pricebookPrices map[string]*float64, pbCurrency string) (price float64, currency string, ok bool) {
	if p, exists := pricebookPrices[strings.TrimSpace(sku.ID)]; exists && p != nil {
		return *p, strings.TrimSpace(pbCurrency), true
	}
	return extractPriceFromDefaultValues(sku.DefaultValues)
}

func (s *Service) evalSKUItems(
	tx *gorm.DB,
	tenantUUID string,
	skus []skuRow,
	channelCfgBySPU map[string]*productmodel.ChannelVisibility,
) ([]ItemDTO, error) {
	skuIDs := make([]string, 0, len(skus))
	for _, sku := range skus {
		if strings.TrimSpace(sku.ID) == "" {
			continue
		}
		skuIDs = append(skuIDs, sku.ID)
	}

	priceBySKU, pbCurrency := s.loadBasePricebookSkuPrices(tx, tenantUUID, skuIDs)
	stockBySKU := s.loadAvailableQty(tx, tenantUUID, skuIDs)

	items := make([]ItemDTO, 0, len(skus))
	for _, sku := range skus {
		reasons := make([]string, 0, 4)

		cfg := (*productmodel.ChannelVisibility)(nil)
		if channelCfgBySPU != nil {
			cfg = channelCfgBySPU[strings.TrimSpace(sku.SPUID)]
		}
		reasons = append(reasons, s.evalChannelGate(cfg)...)

		if !isMiniAppSKUOnline(sku.Status) {
			reasons = appendReason(reasons, string(ReasonSKUNotOnline))
		}

		price, currency, ok := s.resolveSKUPrice(sku, priceBySKU, pbCurrency)
		var priceDTO *MoneyDTO
		if ok && price > 0 {
			priceDTO = &MoneyDTO{Amount: price, Currency: currency}
		} else {
			reasons = appendReason(reasons, string(ReasonNoPublicPrice))
		}

		available := 0
		if v, ok := stockBySKU[sku.ID]; ok && v > 0 {
			available = v
		} else {
			reasons = appendReason(reasons, string(ReasonOutOfStock))
		}

		items = append(items, ItemDTO{
			SKUID:        sku.ID,
			Sellable:     len(reasons) == 0,
			Reasons:      reasons,
			Price:        priceDTO,
			AvailableQty: available,
		})
	}
	return items, nil
}

func extractPriceFromDefaultValues(raw datatypes.JSON) (price float64, currency string, ok bool) {
	if len(raw) == 0 {
		return 0, "", false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, "", false
	}
	currency = strings.ToUpper(strings.TrimSpace(asString(m["currency"])))
	if currency == "" {
		currency = "CNY"
	}
	if v, ok2 := asFloat(m["salePrice"]); ok2 {
		return v, currency, true
	}
	if v, ok2 := asFloat(m["basePrice"]); ok2 {
		return v, currency, true
	}
	if v, ok2 := asFloat(m["msrp"]); ok2 {
		return v, currency, true
	}
	return 0, currency, false
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func isMiniAppSKUOnline(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	_, ok := miniAppSellableSKUStatuses[status]
	return ok
}

func appendReason(reasons []string, reason string) []string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return reasons
	}
	for _, r := range reasons {
		if r == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}
