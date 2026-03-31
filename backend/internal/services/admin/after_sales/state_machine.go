package after_sales

import (
	"fmt"
	"strings"
)

const (
	CaseStatusPending   = "pending"
	CaseStatusAccepted  = "accepted"
	CaseStatusReviewing = "reviewing"
	CaseStatusApproved  = "approved"
	CaseStatusRejected  = "rejected"
	CaseStatusCompleted = "completed"
	CaseStatusClosed    = "closed"
)

var caseStateTransitions = map[string][]string{
	CaseStatusPending:   {CaseStatusAccepted},
	CaseStatusAccepted:  {CaseStatusReviewing},
	CaseStatusReviewing: {CaseStatusApproved, CaseStatusRejected},
	CaseStatusApproved:  {CaseStatusCompleted, CaseStatusClosed},
	CaseStatusRejected:  {CaseStatusClosed},
	CaseStatusCompleted: {CaseStatusClosed},
}

var activeStatuses = map[string]struct{}{
	CaseStatusPending:   {},
	CaseStatusAccepted:  {},
	CaseStatusReviewing: {},
	CaseStatusApproved:  {},
}

func NormalizeCaseStatus(status string) string {
	v := strings.TrimSpace(strings.ToLower(status))
	switch v {
	case CaseStatusPending, CaseStatusAccepted, CaseStatusReviewing, CaseStatusApproved, CaseStatusRejected, CaseStatusCompleted, CaseStatusClosed:
		return v
	default:
		return ""
	}
}

func ValidateCaseStatusTransition(fromStatus, toStatus string) error {
	from := NormalizeCaseStatus(fromStatus)
	to := NormalizeCaseStatus(toStatus)
	if from == "" || to == "" {
		return fmt.Errorf("invalid status transition: %q -> %q", fromStatus, toStatus)
	}
	if from == to {
		return nil
	}
	for _, candidate := range caseStateTransitions[from] {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition: %s -> %s", from, to)
}

func IsActiveCaseStatus(status string) bool {
	_, ok := activeStatuses[NormalizeCaseStatus(status)]
	return ok
}

func IsCaseTypeAllowsReverseLink(caseType string) bool {
	t := strings.TrimSpace(strings.ToLower(caseType))
	return t == "return_refund" || t == "exchange"
}
