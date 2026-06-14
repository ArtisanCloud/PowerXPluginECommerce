package promotion

import (
	"context"
	"errors"
	"strings"
	"time"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type QuoteService struct {
	deps  *app.Deps
	repo  *promotionrepo.CampaignRepository
	audit *AuditLogService
}

func NewQuoteService(deps *app.Deps) *QuoteService {
	if deps == nil || deps.DB == nil {
		return &QuoteService{deps: deps}
	}
	return &QuoteService{deps: deps, repo: promotionrepo.NewCampaignRepository(deps.DB), audit: NewAuditLogService(deps)}
}

func (s *QuoteService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *QuoteService) Quote(ctx context.Context, in QuoteInput) (*QuoteResult, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	tenantUUID := strings.TrimSpace(in.TenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant uuid is required")
	}
	now := in.SubmittedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	lines, baseTotal, err := quoteLines(in.Items)
	if err != nil {
		return nil, err
	}
	campaigns, err := s.repo.ListActive(ctx, tenantUUID, strings.TrimSpace(in.Channel), now)
	if err != nil {
		return nil, err
	}
	var candidates []quoteCandidate
	var rejected []RejectedPromotion
	for _, campaign := range campaigns {
		applied, stacking, reason := s.evaluateCampaign(campaign, in, baseTotal, now)
		if reason != "" {
			rejected = append(rejected, RejectedPromotion{PromotionID: campaign.ID, Code: campaign.Code, Reason: reason})
			continue
		}
		candidates = append(candidates, quoteCandidate{Applied: applied, Stacking: stacking})
	}
	applied, stackingRejected := selectApplicablePromotions(candidates)
	rejected = append(rejected, stackingRejected...)
	discount := int64(0)
	couponAllowed := true
	for _, item := range applied {
		discount += item.DiscountMinor
		if !item.StackableWithCoupon {
			couponAllowed = false
		}
	}
	if discount > baseTotal {
		discount = baseTotal
	}
	allocations := AllocateDiscountByPrePromotionAmount(lines, discount)
	if strings.TrimSpace(in.Currency) == "" {
		in.Currency = "CNY"
	}
	result := &QuoteResult{
		Currency: strings.TrimSpace(in.Currency), BaseTotalMinor: baseTotal, PromotionDiscountMinor: discount,
		AfterPromotionTotalMinor: baseTotal - discount, CouponStackingAllowed: couponAllowed,
		AppliedPromotions: applied, RejectedPromotions: rejected, LineAllocations: allocations, PricedAt: now,
	}
	return result, nil
}

func (s *QuoteService) evaluateCampaign(campaign promotionmodel.Campaign, in QuoteInput, baseTotal int64, at time.Time) (AppliedPromotion, StackingRule, string) {
	if campaign.Status != promotionmodel.StatusActive {
		return AppliedPromotion{}, StackingRule{}, ReasonPaused
	}
	if at.Before(campaign.ValidFrom) {
		return AppliedPromotion{}, StackingRule{}, ReasonNotStarted
	}
	if at.After(campaign.ValidTo) {
		return AppliedPromotion{}, StackingRule{}, ReasonExpired
	}
	condition := ValueFromJSON(campaign.ConditionRule, ConditionRule{})
	scope := ValueFromJSON(campaign.ScopeRule, ScopeRule{ScopeType: "all"})
	action := ValueFromJSON(campaign.ActionRule, ActionRule{})
	stacking := ValueFromJSON(campaign.StackingRule, StackingRule{Priority: 100, Stackable: true, StackableWithCoupon: true})
	if condition.MinOrderAmountMinor > 0 && baseTotal < condition.MinOrderAmountMinor {
		return AppliedPromotion{}, stacking, ReasonThresholdNotMet
	}
	if !scopeMatches(scope, in) {
		return AppliedPromotion{}, stacking, ReasonScopeMismatch
	}
	discount := calculateDiscount(campaign.PromotionType, action, baseTotal)
	if discount <= 0 {
		return AppliedPromotion{}, stacking, ReasonInvalidRule
	}
	if discount > baseTotal {
		discount = baseTotal
	}
	return AppliedPromotion{
		PromotionID: campaign.ID, Code: campaign.Code, Name: campaign.Name, PromotionType: campaign.PromotionType,
		DiscountMinor: discount, Priority: stacking.Priority, ExclusionGroup: stacking.ExclusionGroup,
		StackableWithCoupon: stacking.StackableWithCoupon,
	}, stacking, ""
}

func quoteLines(items []QuoteItemInput) ([]LineAmount, int64, error) {
	if len(items) == 0 {
		return nil, 0, errors.New("items are required")
	}
	lines := make([]LineAmount, 0, len(items))
	total := int64(0)
	for _, item := range items {
		if strings.TrimSpace(item.SKUID) == "" || item.Qty <= 0 || item.UnitPriceMinor < 0 {
			return nil, 0, errors.New("invalid quote item")
		}
		amount := item.Qty * item.UnitPriceMinor
		total += amount
		lines = append(lines, LineAmount{LineID: strings.TrimSpace(item.LineID), SKUID: strings.TrimSpace(item.SKUID), BaseMinor: amount})
	}
	return lines, total, nil
}

func scopeMatches(scope ScopeRule, in QuoteInput) bool {
	channels := nonEmptyStrings(scope.Channels)
	if len(channels) > 0 {
		ok := false
		for _, ch := range channels {
			if strings.EqualFold(ch, strings.TrimSpace(in.Channel)) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	scopeType := strings.TrimSpace(scope.ScopeType)
	if scopeType == "" || scopeType == "all" {
		return true
	}
	if scopeType != "sku" {
		return false
	}
	allowed := map[string]struct{}{}
	for _, sku := range nonEmptyStrings(scope.SKUIDs) {
		allowed[sku] = struct{}{}
	}
	for _, item := range in.Items {
		if _, ok := allowed[strings.TrimSpace(item.SKUID)]; ok {
			return true
		}
	}
	return false
}

func calculateDiscount(promotionType string, action ActionRule, baseTotal int64) int64 {
	switch strings.TrimSpace(promotionType) {
	case promotionmodel.TypeAmountOff:
		return action.DiscountAmountMinor
	case promotionmodel.TypePercentOff:
		if action.DiscountPercentBps <= 0 || action.DiscountPercentBps > 10000 {
			return 0
		}
		discount := baseTotal - (baseTotal*int64(action.DiscountPercentBps))/10000
		if action.MaxDiscountMinor > 0 && discount > action.MaxDiscountMinor {
			return action.MaxDiscountMinor
		}
		return discount
	default:
		return 0
	}
}
