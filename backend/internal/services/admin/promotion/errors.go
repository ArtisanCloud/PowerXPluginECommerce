package promotion

import "fmt"

const (
	ReasonNotStarted              = "not_started"
	ReasonExpired                 = "expired"
	ReasonPaused                  = "paused"
	ReasonThresholdNotMet         = "threshold_not_met"
	ReasonScopeMismatch           = "scope_mismatch"
	ReasonExclusiveConflict       = "exclusive_conflict"
	ReasonNotStackable            = "not_stackable"
	ReasonInvalidRule             = "invalid_rule"
	ReasonPromotionExcludesCoupon = "promotion_excludes_coupon"
)

type RuleError struct {
	Reason  string
	Message string
}

func (e *RuleError) Error() string {
	if e == nil {
		return "promotion rule error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return fmt.Sprintf("promotion rule failed: %s", e.Reason)
	}
	return "promotion rule failed"
}

func NewRuleError(reason, message string) *RuleError {
	return &RuleError{Reason: reason, Message: message}
}
