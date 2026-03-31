package subscription_reconciliation

import "testing"

func TestClassifyDelta(t *testing.T) {
	tests := []struct {
		name      string
		sample    ReconciliationSample
		wantType  string
		wantRisk  string
		wantDelta int64
	}{
		{name: "missing_payment", sample: ReconciliationSample{ExpectedAmountMinor: 100, ActualAmountMinor: 0}, wantType: "missing_payment", wantRisk: "high", wantDelta: 100},
		{name: "duplicate_payment", sample: ReconciliationSample{ExpectedAmountMinor: 0, ActualAmountMinor: 50}, wantType: "duplicate_payment", wantRisk: "high", wantDelta: -50},
		{name: "amount_mismatch_medium", sample: ReconciliationSample{ExpectedAmountMinor: 100, ActualAmountMinor: 90}, wantType: "amount_mismatch", wantRisk: "medium", wantDelta: 10},
		{name: "amount_mismatch_high", sample: ReconciliationSample{ExpectedAmountMinor: 100, ActualAmountMinor: 50}, wantType: "amount_mismatch", wantRisk: "high", wantDelta: 50},
		{name: "status_mismatch", sample: ReconciliationSample{ExpectedStatus: "paid", ActualStatus: "failed", ExpectedAmountMinor: 100, ActualAmountMinor: 100}, wantType: "status_mismatch", wantRisk: "medium", wantDelta: 0},
		{name: "data_missing", sample: ReconciliationSample{ForceDataMissing: true, ExpectedAmountMinor: 10, ActualAmountMinor: 0}, wantType: "data_missing", wantRisk: "low", wantDelta: 10},
		{name: "matched", sample: ReconciliationSample{ExpectedAmountMinor: 100, ActualAmountMinor: 100}, wantType: "", wantRisk: "", wantDelta: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotRisk, gotDelta := classifyDelta(tt.sample)
			if gotType != tt.wantType || gotRisk != tt.wantRisk || gotDelta != tt.wantDelta {
				t.Fatalf("classifyDelta() = (%s,%s,%d), want (%s,%s,%d)", gotType, gotRisk, gotDelta, tt.wantType, tt.wantRisk, tt.wantDelta)
			}
		})
	}
}
