package order

import (
	"encoding/json"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
)

func toCouponSummaryDTO(in *couponsvc.QuoteResult) *CouponSummaryDTO {
	if in == nil {
		return nil
	}
	out := &CouponSummaryDTO{
		Currency:           in.Currency,
		BaseTotalMinor:     in.BaseTotalMinor,
		DiscountTotalMinor: in.DiscountTotalMinor,
		PayableTotalMinor:  in.PayableTotalMinor,
		PricedAt:           in.PricedAt,
		LineAllocations:    make([]CouponLineAllocationDTO, 0, len(in.LineAllocations)),
		AppliedCoupons:     make([]CouponAppliedDTO, 0, len(in.AppliedCoupons)),
		RejectedCoupons:    make([]CouponRejectedDTO, 0, len(in.RejectedCoupons)),
	}
	for _, row := range in.LineAllocations {
		out.LineAllocations = append(out.LineAllocations, CouponLineAllocationDTO{
			LineID:        row.LineID,
			BaseMinor:     row.BaseMinor,
			DiscountMinor: row.DiscountMinor,
		})
	}
	for _, row := range in.AppliedCoupons {
		out.AppliedCoupons = append(out.AppliedCoupons, CouponAppliedDTO{
			AssetID:       row.AssetID,
			TemplateID:    row.TemplateID,
			CouponCode:    row.CouponCode,
			DiscountMinor: row.DiscountMinor,
			Level:         row.Level,
		})
	}
	for _, row := range in.RejectedCoupons {
		out.RejectedCoupons = append(out.RejectedCoupons, CouponRejectedDTO{
			AssetID: row.AssetID,
			Reason:  row.Reason,
		})
	}
	return out
}

func couponSummaryFromSnapshotRow(row *couponmodel.OrderCouponSnapshot) *CouponSummaryDTO {
	if row == nil {
		return nil
	}
	out := &CouponSummaryDTO{
		Currency:           row.Currency,
		BaseTotalMinor:     row.BaseTotalMinor,
		DiscountTotalMinor: row.DiscountTotalMinor,
		PayableTotalMinor:  row.PayableTotalMinor,
		PricedAt:           row.PricedAt,
		LineAllocations:    []CouponLineAllocationDTO{},
		AppliedCoupons:     []CouponAppliedDTO{},
		RejectedCoupons:    []CouponRejectedDTO{},
	}
	_ = json.Unmarshal(row.LineAllocations, &out.LineAllocations)
	_ = json.Unmarshal(row.AppliedCoupons, &out.AppliedCoupons)
	_ = json.Unmarshal(row.RejectedCoupons, &out.RejectedCoupons)
	return out
}
