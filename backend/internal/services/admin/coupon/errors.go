package coupon

import "fmt"

const (
	ReasonThresholdNotMet = "threshold_not_met"
	ReasonScopeMismatch   = "scope_mismatch"
	ReasonExpired         = "expired"
	ReasonOccupied        = "occupied"
	ReasonNotStackable    = "not_stackable"
	ReasonNotFound        = "coupon_not_found"
	ReasonInvalid         = "coupon_invalid"
)

type RuleError struct {
	Reason  string
	Message string
}

func (e *RuleError) Error() string {
	if e == nil {
		return "coupon rule error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return fmt.Sprintf("coupon rule failed: %s", e.Reason)
	}
	return "coupon rule failed"
}

func NewRuleError(reason, message string) *RuleError {
	return &RuleError{Reason: reason, Message: message}
}
