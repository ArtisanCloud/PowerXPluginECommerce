package promotion

// AlertRule documents promotion alert thresholds consumed by deployment tooling.
type AlertRule struct {
	Name        string
	Description string
	Threshold   string
}

func DefaultAlertRules() []AlertRule {
	return []AlertRule{
		{Name: "promotion_quote_error_rate_high", Description: "促销试算错误率过高", Threshold: "5m error rate > 1%"},
		{Name: "promotion_rejection_spike", Description: "促销拒绝原因突增", Threshold: "5m rejected_total increases sharply"},
	}
}
