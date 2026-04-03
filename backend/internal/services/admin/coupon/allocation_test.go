package coupon

import "testing"

func TestAllocateDiscountByAmount_ProportionalAndExact(t *testing.T) {
	alloc := AllocateDiscountByAmount([]LineAmount{
		{LineID: "l1", BaseMinor: 100, DiscountCap: 100},
		{LineID: "l2", BaseMinor: 200, DiscountCap: 200},
		{LineID: "l3", BaseMinor: 700, DiscountCap: 700},
	}, 101)
	if len(alloc) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(alloc))
	}
	sum := int64(0)
	for _, row := range alloc {
		sum += row.DiscountMinor
		if row.DiscountMinor < 0 || row.DiscountMinor > row.BaseMinor {
			t.Fatalf("invalid discount for %s: %d", row.LineID, row.DiscountMinor)
		}
	}
	if sum != 101 {
		t.Fatalf("expected total discount 101, got %d", sum)
	}
	if alloc[2].DiscountMinor <= alloc[1].DiscountMinor {
		t.Fatalf("expected largest line to get most discount, got %+v", alloc)
	}
}
