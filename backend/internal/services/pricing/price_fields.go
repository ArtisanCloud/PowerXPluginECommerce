package pricing

import "errors"

// PickSettlementPrice returns the effective sale price and the source field name.
// Priority: sale > base > msrp.
func PickSettlementPrice(sale, base, msrp *int64) (amountMinor int64, sourceField string, ok bool) {
	if sale != nil {
		return *sale, "sale_amount_minor", true
	}
	if base != nil {
		return *base, "base_amount_minor", true
	}
	if msrp != nil {
		return *msrp, "msrp_amount_minor", true
	}
	return 0, "", false
}

func validateNonNegative(name string, v *int64) error {
	if v == nil {
		return nil
	}
	if *v < 0 {
		return errors.New(name + " must be >= 0")
	}
	return nil
}
