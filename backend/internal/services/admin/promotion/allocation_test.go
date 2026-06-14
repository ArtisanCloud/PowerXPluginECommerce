package promotion

import "testing"

func TestAllocateDiscountByPrePromotionAmount(t *testing.T) {
	out := AllocateDiscountByPrePromotionAmount([]LineAmount{
		{LineID: "l1", SKUID: "sku-1", BaseMinor: 1000},
		{LineID: "l2", SKUID: "sku-2", BaseMinor: 500},
	}, 300)
	if len(out) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(out))
	}
	if out[0].PromotionDiscountMinor+out[1].PromotionDiscountMinor != 300 {
		t.Fatalf("allocation sum mismatch: %+v", out)
	}
	if out[0].PromotionDiscountMinor != 200 || out[1].PromotionDiscountMinor != 100 {
		t.Fatalf("unexpected proportional allocation: %+v", out)
	}
}
