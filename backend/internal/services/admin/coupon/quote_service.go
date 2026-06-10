package coupon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type QuoteItemInput struct {
	LineID         string `json:"line_id"`
	SKUID          string `json:"sku_id"`
	Qty            int64  `json:"qty"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
}

type QuoteInput struct {
	TenantUUID string
	UserID     string
	Channel    string
	CouponIDs  []string
	Items      []QuoteItemInput
	Now        time.Time
	Currency   string
}

type AppliedCoupon struct {
	AssetID       string `json:"asset_id"`
	TemplateID    string `json:"template_id"`
	CouponCode    string `json:"coupon_code"`
	DiscountMinor int64  `json:"discount_minor"`
	Level         string `json:"level"`
}

type RejectedCoupon struct {
	AssetID string `json:"asset_id"`
	Reason  string `json:"reason"`
}

type QuoteResult struct {
	Currency           string           `json:"currency"`
	BaseTotalMinor     int64            `json:"base_total_minor"`
	DiscountTotalMinor int64            `json:"discount_total_minor"`
	PayableTotalMinor  int64            `json:"payable_total_minor"`
	LineAllocations    []LineAllocation `json:"line_allocations"`
	AppliedCoupons     []AppliedCoupon  `json:"applied_coupons"`
	RejectedCoupons    []RejectedCoupon `json:"rejected_coupons"`
	PricedAt           time.Time        `json:"priced_at"`
}

type QuoteService struct {
	templateRepo *couponrepo.TemplateRepository
	assetRepo    *couponrepo.AssetRepository
}

func NewQuoteService(deps *app.Deps) *QuoteService {
	if deps == nil || deps.DB == nil {
		return &QuoteService{}
	}
	return &QuoteService{
		templateRepo: couponrepo.NewTemplateRepository(deps.DB),
		assetRepo:    couponrepo.NewAssetRepository(deps.DB),
	}
}

func (s *QuoteService) Ready() bool {
	return s != nil && s.templateRepo != nil && s.assetRepo != nil
}

func (s *QuoteService) Quote(ctx context.Context, in QuoteInput) (*QuoteResult, error) {
	if !s.Ready() {
		return nil, errors.New("coupon quote service unavailable")
	}
	in.TenantUUID = strings.TrimSpace(in.TenantUUID)
	in.UserID = strings.TrimSpace(in.UserID)
	if in.TenantUUID == "" {
		return nil, errors.New("tenant uuid is required")
	}
	if in.UserID == "" {
		return nil, errors.New("user id is required")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("quote items are required")
	}
	if in.Now.IsZero() {
		in.Now = time.Now().UTC()
	}
	if strings.TrimSpace(in.Currency) == "" {
		in.Currency = "CNY"
	}

	lineBase := map[string]int64{}
	lineOrder := make([]string, 0, len(in.Items))
	baseTotal := int64(0)
	for _, item := range in.Items {
		skuID := strings.TrimSpace(item.SKUID)
		if skuID == "" || item.Qty <= 0 || item.UnitPriceMinor < 0 {
			return nil, errors.New("invalid quote item")
		}
		lineID := strings.TrimSpace(item.LineID)
		if lineID == "" {
			lineID = skuID
		}
		amount := item.Qty * item.UnitPriceMinor
		lineOrder = append(lineOrder, lineID)
		lineBase[lineID] = amount
		baseTotal += amount
	}
	linePayable := make(map[string]int64, len(lineBase))
	for k, v := range lineBase {
		linePayable[k] = v
	}

	result := &QuoteResult{
		Currency:          in.Currency,
		BaseTotalMinor:    baseTotal,
		PayableTotalMinor: baseTotal,
		PricedAt:          in.Now.UTC(),
		AppliedCoupons:    []AppliedCoupon{},
		RejectedCoupons:   []RejectedCoupon{},
	}
	if len(in.CouponIDs) == 0 {
		result.LineAllocations = buildLineAllocations(lineOrder, lineBase, map[string]int64{})
		return result, nil
	}

	assets, err := s.assetRepo.ListAvailableByUserAndIDs(ctx, in.TenantUUID, in.UserID, in.CouponIDs, in.Now)
	if err != nil {
		return nil, err
	}
	assetByID := make(map[string]couponmodel.CouponAsset, len(assets))
	templateIDs := make([]string, 0, len(assets))
	for _, asset := range assets {
		assetByID[strings.TrimSpace(asset.ID)] = asset
		templateIDs = append(templateIDs, asset.TemplateID)
	}
	for _, id := range in.CouponIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := assetByID[id]; !ok {
			result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: id, Reason: ReasonNotFound})
		}
	}

	templateByID, err := s.templateRepo.MapByID(ctx, in.TenantUUID, templateIDs, in.Now)
	if err != nil {
		return nil, err
	}

	metas := make([]StackingMeta, 0, len(assets))
	for _, asset := range assets {
		tpl, ok := templateByID[asset.TemplateID]
		if !ok {
			result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: asset.ID, Reason: ReasonInvalid})
			continue
		}
		meta := parseStackingMeta(asset.ID, tpl)
		metas = append(metas, meta)
	}

	ordered, rejected := SortAndFilterStackable(metas)
	for assetID, reason := range rejected {
		result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: assetID, Reason: reason})
	}

	lineDiscount := map[string]int64{}
	for _, meta := range ordered {
		asset, ok := assetByID[meta.AssetID]
		if !ok {
			continue
		}
		tpl, ok := templateByID[asset.TemplateID]
		if !ok {
			result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: asset.ID, Reason: ReasonInvalid})
			continue
		}
		alloc, reason := calculateDiscountAllocations(tpl, in.Items, linePayable)
		if len(alloc) == 0 {
			if reason == "" {
				reason = ReasonScopeMismatch
			}
			result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: asset.ID, Reason: reason})
			continue
		}
		discount := int64(0)
		for lineID, d := range alloc {
			if d <= 0 {
				continue
			}
			lineDiscount[lineID] += d
			linePayable[lineID] -= d
			discount += d
		}
		if discount <= 0 {
			result.RejectedCoupons = append(result.RejectedCoupons, RejectedCoupon{AssetID: asset.ID, Reason: ReasonThresholdNotMet})
			continue
		}
		result.AppliedCoupons = append(result.AppliedCoupons, AppliedCoupon{
			AssetID:       asset.ID,
			TemplateID:    tpl.ID,
			CouponCode:    asset.CouponCode,
			DiscountMinor: discount,
			Level:         meta.Level,
		})
	}

	result.LineAllocations = buildLineAllocations(lineOrder, lineBase, lineDiscount)
	for _, line := range result.LineAllocations {
		result.DiscountTotalMinor += line.DiscountMinor
	}
	result.PayableTotalMinor = result.BaseTotalMinor - result.DiscountTotalMinor
	if result.PayableTotalMinor < 0 {
		result.PayableTotalMinor = 0
	}
	return result, nil
}

func buildLineAllocations(lineOrder []string, lineBase, lineDiscount map[string]int64) []LineAllocation {
	alloc := make([]LineAllocation, 0, len(lineOrder))
	for _, lineID := range lineOrder {
		alloc = append(alloc, LineAllocation{
			LineID:        lineID,
			BaseMinor:     lineBase[lineID],
			DiscountMinor: lineDiscount[lineID],
		})
	}
	return alloc
}

func parseStackingMeta(assetID string, tpl couponmodel.CouponTemplate) StackingMeta {
	meta := StackingMeta{AssetID: assetID, Level: StackLevelOrder, Priority: 100, Stackable: true}
	var rule map[string]any
	_ = json.Unmarshal(tpl.StackingRule, &rule)
	if v, ok := rule["priority"].(float64); ok {
		meta.Priority = int(v)
	}
	if v, ok := rule["stackable"].(bool); ok {
		meta.Stackable = v
	}
	if v, ok := rule["exclusion_group"].(string); ok {
		meta.ExclusionGroup = strings.TrimSpace(v)
	}
	if v, ok := rule["level"].(string); ok {
		level := strings.ToLower(strings.TrimSpace(v))
		if level == StackLevelItem || level == StackLevelOrder {
			meta.Level = level
		}
	}
	if meta.Level == StackLevelOrder {
		var scope map[string]any
		_ = json.Unmarshal(tpl.ScopeRule, &scope)
		if _, ok := scope["sku_ids"]; ok {
			meta.Level = StackLevelItem
		}
	}
	return meta
}

func calculateDiscountAllocations(tpl couponmodel.CouponTemplate, items []QuoteItemInput, linePayable map[string]int64) (map[string]int64, string) {
	var threshold, scope map[string]any
	_ = json.Unmarshal(tpl.ThresholdRule, &threshold)
	_ = json.Unmarshal(tpl.ScopeRule, &scope)

	scopeSet := map[string]struct{}{}
	if arr, ok := scope["sku_ids"].([]any); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				scopeSet[strings.TrimSpace(s)] = struct{}{}
			}
		}
	}
	level := StackLevelOrder
	if v, ok := scope["level"].(string); ok {
		lv := strings.ToLower(strings.TrimSpace(v))
		if lv == StackLevelItem || lv == StackLevelOrder {
			level = lv
		}
	}
	if len(scopeSet) > 0 {
		level = StackLevelItem
	}

	eligibleLines := make([]LineAmount, 0, len(items))
	for _, item := range items {
		lineID := strings.TrimSpace(item.LineID)
		if lineID == "" {
			lineID = strings.TrimSpace(item.SKUID)
		}
		if lineID == "" {
			continue
		}
		if len(scopeSet) > 0 {
			if _, ok := scopeSet[strings.TrimSpace(item.SKUID)]; !ok {
				continue
			}
		}
		payable := linePayable[lineID]
		if payable <= 0 {
			continue
		}
		eligibleLines = append(eligibleLines, LineAmount{LineID: lineID, BaseMinor: payable, DiscountCap: payable})
	}
	if len(eligibleLines) == 0 {
		return nil, ReasonScopeMismatch
	}

	eligibleTotal := int64(0)
	for _, line := range eligibleLines {
		eligibleTotal += line.BaseMinor
	}
	minAmount := int64(0)
	if v, ok := threshold["min_amount_minor"].(float64); ok {
		minAmount = int64(v)
	}
	if minAmount > 0 && eligibleTotal < minAmount {
		return nil, ReasonThresholdNotMet
	}

	discount := int64(0)
	typ := strings.ToLower(strings.TrimSpace(tpl.CouponType))
	switch typ {
	case "amount":
		if v, ok := threshold["amount_minor"].(float64); ok {
			discount = int64(v)
		} else {
			var scopeAmount map[string]any
			_ = json.Unmarshal(tpl.ScopeRule, &scopeAmount)
			if v2, ok := scopeAmount["amount_minor"].(float64); ok {
				discount = int64(v2)
			}
		}
	case "percent":
		percentBps := int64(0)
		if v, ok := threshold["percent_bps"].(float64); ok {
			percentBps = int64(v)
		}
		if percentBps == 0 {
			if v, ok := threshold["percent"].(float64); ok {
				if v <= 1 {
					percentBps = int64(v * 10000)
				} else {
					percentBps = int64(v * 100)
				}
			}
		}
		discount = int64(math.Floor(float64(eligibleTotal*percentBps) / 10000.0))
	default:
		return nil, ReasonInvalid
	}
	if discount <= 0 {
		return nil, ReasonThresholdNotMet
	}
	if discount > eligibleTotal {
		discount = eligibleTotal
	}

	allocated := AllocateDiscountByAmount(eligibleLines, discount)
	out := make(map[string]int64, len(allocated))
	for _, line := range allocated {
		if line.DiscountMinor > 0 {
			out[line.LineID] = line.DiscountMinor
		}
	}
	if len(out) == 0 {
		return nil, ReasonThresholdNotMet
	}
	if level == StackLevelOrder {
		return out, ""
	}
	return out, ""
}

func SumAppliedDiscount(applied []AppliedCoupon) int64 {
	var sum int64
	for _, row := range applied {
		sum += row.DiscountMinor
	}
	return sum
}

func buildReserveIdempotencyKey(orderID, assetID string) string {
	return fmt.Sprintf("reserve:%s:%s", strings.TrimSpace(orderID), strings.TrimSpace(assetID))
}
