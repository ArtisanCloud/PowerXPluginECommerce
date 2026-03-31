package subscription_reconciliation

import (
	"strings"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

type governanceCandidate struct {
	SubscriptionRef string
	ReasonCode      string
}

type candidateResult struct {
	retrySucceeded bool
	retryFailed    int
	logs           []SubscriptionReconciliationModel.RenewalExecutionLog
}

func buildGovernanceCandidates(deltas []SubscriptionReconciliationModel.ReconciliationDelta) []governanceCandidate {
	candidates := make([]governanceCandidate, 0, len(deltas))
	seen := map[string]struct{}{}
	for _, delta := range deltas {
		ref := strings.TrimSpace(delta.SubscriptionRef)
		if ref == "" {
			continue
		}
		if delta.Status == "resolved" || delta.Status == "ignored" {
			continue
		}
		if delta.DeltaType != "missing_payment" && delta.DeltaType != "status_mismatch" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		candidates = append(candidates, governanceCandidate{
			SubscriptionRef: ref,
			ReasonCode:      strings.TrimSpace(delta.ReasonCode),
		})
	}
	return candidates
}

func executeCandidateRetries(now time.Time, cycleStart time.Time, candidate governanceCandidate, policy governancePolicy) candidateResult {
	result := candidateResult{
		logs: make([]SubscriptionReconciliationModel.RenewalExecutionLog, 0, len(policy.retryWindows)),
	}
	for i, window := range policy.retryWindows {
		attempt := i + 1
		scheduledAt := cycleStart.Add(window)
		retrySuccess := shouldRetrySucceed(candidate.ReasonCode, attempt)
		log := SubscriptionReconciliationModel.RenewalExecutionLog{
			ID:              utils.NewUUID(),
			SubscriptionRef: candidate.SubscriptionRef,
			ActionType:      "retry",
			AttemptNo:       attempt,
			ScheduledAt:     &scheduledAt,
			ExecutedAt:      &now,
			Result:          "failed",
			OperatorType:    "system",
		}
		if retrySuccess {
			log.Result = "success"
			result.retrySucceeded = true
			result.logs = append(result.logs, log)
			return result
		}
		log.FailureReason = "payment_declined"
		result.retryFailed++
		result.logs = append(result.logs, log)
	}
	return result
}

func shouldRetrySucceed(reasonCode string, attempt int) bool {
	switch strings.TrimSpace(strings.ToLower(reasonCode)) {
	case "success_at_1":
		return attempt >= 1
	case "success_at_2":
		return attempt >= 2
	case "success_at_3":
		return attempt >= 3
	case "success_at_4":
		return attempt >= 4
	case "always_fail":
		return false
	default:
		return attempt >= 2
	}
}
