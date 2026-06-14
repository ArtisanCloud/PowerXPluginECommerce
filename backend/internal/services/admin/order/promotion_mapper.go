package order

import (
	"encoding/json"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
)

func toPromotionSummaryDTO(in *promotionsvc.QuoteResult) *PromotionSummaryDTO {
	if in == nil {
		return nil
	}
	out := &PromotionSummaryDTO{
		Currency:                 in.Currency,
		BaseTotalMinor:           in.BaseTotalMinor,
		PromotionDiscountMinor:   in.PromotionDiscountMinor,
		AfterPromotionTotalMinor: in.AfterPromotionTotalMinor,
		CouponStackingAllowed:    in.CouponStackingAllowed,
		PricedAt:                 in.PricedAt,
		LineAllocations:          make([]PromotionLineAllocationDTO, 0, len(in.LineAllocations)),
		AppliedPromotions:        make([]PromotionAppliedDTO, 0, len(in.AppliedPromotions)),
		RejectedPromotions:       make([]PromotionRejectedDTO, 0, len(in.RejectedPromotions)),
	}
	for _, row := range in.LineAllocations {
		out.LineAllocations = append(out.LineAllocations, PromotionLineAllocationDTO(row))
	}
	for _, row := range in.AppliedPromotions {
		out.AppliedPromotions = append(out.AppliedPromotions, PromotionAppliedDTO(row))
	}
	for _, row := range in.RejectedPromotions {
		out.RejectedPromotions = append(out.RejectedPromotions, PromotionRejectedDTO(row))
	}
	return out
}

func promotionSummaryFromSnapshotRow(row *promotionmodel.OrderSnapshot) *PromotionSummaryDTO {
	if row == nil {
		return nil
	}
	out := &PromotionSummaryDTO{
		Currency:                 row.Currency,
		BaseTotalMinor:           row.BaseTotalMinor,
		PromotionDiscountMinor:   row.PromotionDiscountMinor,
		AfterPromotionTotalMinor: row.AfterPromotionTotalMinor,
		LineAllocations:          []PromotionLineAllocationDTO{},
		AppliedPromotions:        []PromotionAppliedDTO{},
		RejectedPromotions:       []PromotionRejectedDTO{},
		PricedAt:                 row.PricedAt,
	}
	_ = json.Unmarshal(row.LineAllocations, &out.LineAllocations)
	_ = json.Unmarshal(row.AppliedPromotions, &out.AppliedPromotions)
	_ = json.Unmarshal(row.RejectedPromotions, &out.RejectedPromotions)
	for _, item := range out.AppliedPromotions {
		if !item.StackableWithCoupon {
			out.CouponStackingAllowed = false
			return out
		}
	}
	out.CouponStackingAllowed = true
	return out
}
