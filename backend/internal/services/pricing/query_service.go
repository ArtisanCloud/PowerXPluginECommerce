package pricing

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	pricingRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

const (
	NoPriceScopeMismatch = "SCOPE_MISMATCH"
	NoPriceNoCandidate   = "NO_ACTIVE_VERSION"
	NoPriceItemMissing   = "ITEM_MISSING"
	NoPriceNoPriceFields = "NO_PRICE_FIELDS"
)

const (
	FallbackNone             = "none"
	FallbackBaseItemFallback = "base_item_fallback"
)

type QueryService struct {
	*Service
}

func NewQueryService(deps *app.Deps) *QueryService {
	svc := NewService(deps)
	if deps == nil || deps.DB == nil {
		return &QueryService{Service: svc}
	}
	return &QueryService{Service: svc}
}

type QueryInput struct {
	SKUID           string
	Currency        string
	ChannelID       *string
	CustomerGroupID *string
	SupplierID      *string
	AsOf            *time.Time
}

type QueryResult struct {
	Currency    string
	Priced      bool
	SourceField *string
	AmountMinor *int64
	Matched     *Matched
	Fields      map[string]any
	Trace       Trace
}

type Matched struct {
	PricebookID   string
	PricebookCode string
	VersionID     string
	Version       int
	SKUID         string
}

type Trace struct {
	Fallback      string
	NoPriceReason *string
}

func (s *QueryService) Query(ctx context.Context, in QueryInput) (*QueryResult, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	if pbSvc := NewPricebookService(s.deps); pbSvc != nil {
		if _, err := pbSvc.EnsureBasePricebook(ctx, strings.TrimSpace(in.Currency), defaultActorFromTenant(tenantUUID)); err != nil {
			return nil, err
		}
	}

	skuID := strings.TrimSpace(in.SKUID)
	currency := strings.TrimSpace(in.Currency)
	if skuID == "" || currency == "" {
		return nil, E(CodeInvalidArgument, errors.New("sku_id and currency are required"))
	}

	asOf := time.Now().UTC()
	if in.AsOf != nil {
		asOf = in.AsOf.UTC()
	}

	tx, err := s.PricebookRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	repo := pricingRepo.NewPriceQueryRepository(tx)

	candidates, err := repo.ListActiveCandidates(ctx, currency, asOf)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		reason := NoPriceNoCandidate
		return &QueryResult{
			Currency: currency,
			Priced:   false,
			Fields:   map[string]any{},
			Trace:    Trace{Fallback: FallbackNone, NoPriceReason: &reason},
		}, nil
	}

	best, bestSpecificity, scopeReason, err := pickBestCandidate(ctx, repo, candidates, in)
	if err != nil {
		return nil, err
	}
	if best == nil {
		reason := scopeReason
		return &QueryResult{
			Currency: currency,
			Priced:   false,
			Fields:   map[string]any{},
			Trace:    Trace{Fallback: FallbackNone, NoPriceReason: &reason},
		}, nil
	}
	_ = bestSpecificity

	quote, priced, err := s.priceFromVersion(ctx, repo, best.Pricebook, best.Version, skuID, FallbackNone)
	if err != nil {
		return nil, err
	}
	if priced {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return quote, nil
	}

	// fallback only when a version is matched but SKU item missing.
	if quote.Trace.NoPriceReason == nil || *quote.Trace.NoPriceReason != NoPriceItemMissing {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return quote, nil
	}

	if strings.EqualFold(best.Pricebook.Code, BasePricebookCode) {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return quote, nil
	}

	baseCandidate, _, _, err := pickBestCandidate(ctx, repo, filterCandidatesByCode(candidates, BasePricebookCode), in)
	if err != nil {
		return nil, err
	}
	if baseCandidate == nil {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return quote, nil
	}

	fallbackQuote, fallbackPriced, err := s.priceFromVersion(ctx, repo, baseCandidate.Pricebook, baseCandidate.Version, skuID, FallbackBaseItemFallback)
	if err != nil {
		return nil, err
	}
	if fallbackPriced {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return fallbackQuote, nil
	}

	// fallback failed => keep original no price reason.
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return quote, nil
}

func (s *QueryService) priceFromVersion(
	ctx context.Context,
	repo *pricingRepo.PriceQueryRepository,
	pb pricingModel.Pricebook,
	v pricingModel.PricebookVersion,
	skuID string,
	fallback string,
) (*QueryResult, bool, error) {
	item, err := repo.FindItem(ctx, v.ID, skuID)
	if err != nil {
		return nil, false, err
	}
	if item == nil {
		reason := NoPriceItemMissing
		return &QueryResult{
			Currency: pb.Currency,
			Priced:   false,
			Fields:   map[string]any{},
			Trace:    Trace{Fallback: fallback, NoPriceReason: &reason},
		}, false, nil
	}

	fields := map[string]any{
		"base_amount_minor": item.BaseAmount,
		"sale_amount_minor": item.SaleAmount,
		"msrp_amount_minor": item.MsrpAmount,
		"cost_amount_minor": item.CostAmount,
		"min_amount_minor":  item.MinAmount,
		"max_amount_minor":  item.MaxAmount,
		"tax_included":      item.TaxIncluded,
	}

	amountMinor, source, ok := PickSettlementPrice(item.SaleAmount, item.BaseAmount, item.MsrpAmount)
	if !ok {
		reason := NoPriceNoPriceFields
		return &QueryResult{
			Currency: pb.Currency,
			Priced:   false,
			Fields:   fields,
			Trace:    Trace{Fallback: fallback, NoPriceReason: &reason},
		}, false, nil
	}

	sourceField := map[string]string{
		"sale_amount_minor": "sale",
		"base_amount_minor": "base",
		"msrp_amount_minor": "msrp",
	}[source]
	amount := amountMinor
	return &QueryResult{
		Currency:    pb.Currency,
		Priced:      true,
		SourceField: &sourceField,
		AmountMinor: &amount,
		Matched: &Matched{
			PricebookID:   pb.ID,
			PricebookCode: pb.Code,
			VersionID:     v.ID,
			Version:       v.Version,
			SKUID:         skuID,
		},
		Fields: fields,
		Trace:  Trace{Fallback: fallback, NoPriceReason: nil},
	}, true, nil
}

func pickBestCandidate(
	ctx context.Context,
	repo *pricingRepo.PriceQueryRepository,
	candidates []pricingRepo.ActiveCandidate,
	in QueryInput,
) (*pricingRepo.ActiveCandidate, int, string, error) {
	if len(candidates) == 0 {
		return nil, 0, NoPriceNoCandidate, nil
	}
	type scored struct {
		idx         int
		specificity int
		publishedAt time.Time
	}
	var scoredList []scored
	for i := range candidates {
		pb := candidates[i].Pricebook
		scopes, err := repo.ListScopesByPricebook(ctx, pb.ID)
		if err != nil {
			return nil, 0, "", err
		}
		ok, specificity := matchScopes(scopes, in)
		if !ok {
			continue
		}
		publishedAt := time.Time{}
		if candidates[i].Version.PublishedAt != nil {
			publishedAt = candidates[i].Version.PublishedAt.UTC()
		}
		scoredList = append(scoredList, scored{idx: i, specificity: specificity, publishedAt: publishedAt})
	}
	if len(scoredList) == 0 {
		return nil, 0, NoPriceScopeMismatch, nil
	}

	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].specificity != scoredList[j].specificity {
			return scoredList[i].specificity > scoredList[j].specificity
		}
		if !scoredList[i].publishedAt.Equal(scoredList[j].publishedAt) {
			return scoredList[i].publishedAt.After(scoredList[j].publishedAt)
		}
		vi := candidates[scoredList[i].idx].Version
		vj := candidates[scoredList[j].idx].Version
		if vi.Version != vj.Version {
			return vi.Version > vj.Version
		}
		return candidates[scoredList[i].idx].Pricebook.ID < candidates[scoredList[j].idx].Pricebook.ID
	})
	best := candidates[scoredList[0].idx]
	return &best, scoredList[0].specificity, "", nil
}

func matchScopes(scopes []*pricingModel.PricebookScope, in QueryInput) (bool, int) {
	if len(scopes) == 0 {
		return true, 0
	}
	req := map[string]*string{
		"channel":        in.ChannelID,
		"customer_group": in.CustomerGroupID,
		"supplier":       in.SupplierID,
	}
	dims := map[string]map[string]struct{}{}
	for _, s := range scopes {
		if s == nil {
			continue
		}
		d := strings.TrimSpace(s.Dimension)
		id := strings.TrimSpace(s.DimensionID)
		if d == "" || id == "" {
			continue
		}
		if _, ok := dims[d]; !ok {
			dims[d] = map[string]struct{}{}
		}
		dims[d][id] = struct{}{}
	}
	if len(dims) == 0 {
		return true, 0
	}
	for dim, allowed := range dims {
		valPtr, ok := req[dim]
		if !ok {
			return false, 0
		}
		if valPtr == nil || strings.TrimSpace(*valPtr) == "" {
			return false, 0
		}
		if _, hit := allowed[strings.TrimSpace(*valPtr)]; !hit {
			return false, 0
		}
	}
	return true, len(dims)
}

func filterCandidatesByCode(c []pricingRepo.ActiveCandidate, code string) []pricingRepo.ActiveCandidate {
	if len(c) == 0 {
		return nil
	}
	var out []pricingRepo.ActiveCandidate
	for _, cand := range c {
		if strings.EqualFold(strings.TrimSpace(cand.Pricebook.Code), strings.TrimSpace(code)) {
			out = append(out, cand)
		}
	}
	return out
}
