package promotion

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

func TestSnapshotService_SaveAndFindByOrderID(t *testing.T) {
	db := setupPromotionDB(t, "file:promotion_snapshot_service?mode=memory&cache=shared")
	svc := NewSnapshotService(&app.Deps{DB: db})
	result := &QuoteResult{
		Currency: "CNY", BaseTotalMinor: 1000, PromotionDiscountMinor: 100, AfterPromotionTotalMinor: 900,
		CouponStackingAllowed: true,
		AppliedPromotions:     []AppliedPromotion{{PromotionID: "promo-1", Code: "ORDER1000", DiscountMinor: 100}},
		LineAllocations:       []LineAllocation{{LineID: "l1", SKUID: "sku-1", BaseAmountMinor: 1000, PromotionDiscountMinor: 100, AfterPromotionAmountMinor: 900}},
		PricedAt:              time.Now().UTC(),
	}

	row, err := svc.Save(context.Background(), nil, "tenant-1", "order-1", result)
	if err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	if row.PromotionDiscountMinor != 100 {
		t.Fatalf("unexpected discount: %+v", row)
	}

	got, err := svc.FindByOrderID(context.Background(), "tenant-1", "order-1")
	if err != nil {
		t.Fatalf("find snapshot: %v", err)
	}
	if got.AfterPromotionTotalMinor != 900 || len(got.AppliedPromotions) == 0 {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
}
