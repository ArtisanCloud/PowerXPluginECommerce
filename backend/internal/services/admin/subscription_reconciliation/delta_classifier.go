package subscription_reconciliation

import "math"

// ReconciliationSample represents one input line for delta classification.
type ReconciliationSample struct {
	SubscriptionRef     string
	BillRef             string
	PaymentRef          string
	ExpectedAmountMinor int64
	ActualAmountMinor   int64
	ExpectedStatus      string
	ActualStatus        string
	ForceDataMissing    bool
	ReasonCode          string
}

// classifyDelta returns delta type/risk and amount gap.
func classifyDelta(sample ReconciliationSample) (deltaType, riskLevel string, deltaAmount int64) {
	deltaAmount = sample.ExpectedAmountMinor - sample.ActualAmountMinor
	if sample.ForceDataMissing {
		return "data_missing", "low", deltaAmount
	}
	if sample.ExpectedStatus != "" && sample.ActualStatus != "" && sample.ExpectedStatus != sample.ActualStatus {
		return "status_mismatch", "medium", deltaAmount
	}
	if sample.ExpectedAmountMinor > 0 && sample.ActualAmountMinor == 0 {
		return "missing_payment", "high", deltaAmount
	}
	if sample.ExpectedAmountMinor == 0 && sample.ActualAmountMinor > 0 {
		return "duplicate_payment", "high", deltaAmount
	}
	if sample.ExpectedAmountMinor != sample.ActualAmountMinor {
		ratio := 0.0
		if sample.ExpectedAmountMinor != 0 {
			ratio = math.Abs(float64(deltaAmount)) / math.Abs(float64(sample.ExpectedAmountMinor))
		}
		if ratio >= 0.2 {
			return "amount_mismatch", "high", deltaAmount
		}
		return "amount_mismatch", "medium", deltaAmount
	}
	return "", "", 0
}
