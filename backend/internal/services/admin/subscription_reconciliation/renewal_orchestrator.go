package subscription_reconciliation

import (
	"context"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	SubscriptionReconciliationObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/subscription_reconciliation"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
)

func (s *Service) runRenewalGovernance(ctx context.Context, tenantUUID, billingCycle string, dryRun bool) (map[string]any, error) {
	batches, err := s.batchRepo.List(ctx, billingCycle, "completed", 200)
	if err != nil {
		return nil, err
	}
	allDeltas := make([]SubscriptionReconciliationModel.ReconciliationDelta, 0, 64)
	for _, batch := range batches {
		deltas, err := s.deltaRepo.ListByBatchID(ctx, batch.ID, "", "", 500)
		if err != nil {
			return nil, err
		}
		allDeltas = append(allDeltas, deltas...)
	}
	candidates := buildGovernanceCandidates(allDeltas)
	policy, err := s.loadGovernancePolicy(ctx)
	if err != nil {
		return nil, err
	}
	cycleStart, _ := time.Parse("2006-01-02", billingCycle)
	now := time.Now().UTC()

	retrySucceeded := 0
	retryFailed := 0
	escalated := 0
	logs := make([]SubscriptionReconciliationModel.RenewalExecutionLog, 0, len(candidates)*len(policy.retryWindows))
	for _, candidate := range candidates {
		candidateResult := executeCandidateRetries(now, cycleStart, candidate, policy)
		retryFailed += candidateResult.retryFailed
		if candidateResult.retrySucceeded {
			retrySucceeded++
			SubscriptionReconciliationObs.RecordGovernanceRecovery(tenantUUID)
		} else {
			SubscriptionReconciliationObs.RecordGovernanceFailure(tenantUUID)
		}
		logs = append(logs, candidateResult.logs...)
		for _, entry := range candidateResult.logs {
			s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventRenewalRetryExecuted, map[string]any{
				"subscription_ref": entry.SubscriptionRef,
				"attempt_no":       entry.AttemptNo,
				"result":           entry.Result,
			})
		}
		if !candidateResult.retrySucceeded && candidateResult.retryFailed >= policy.escalationThreshold {
			escalated++
			notifyLog := SubscriptionReconciliationModel.RenewalExecutionLog{
				ID:              utils.NewUUID(),
				SubscriptionRef: candidate.SubscriptionRef,
				ActionType:      "notify",
				AttemptNo:       candidateResult.retryFailed,
				ExecutedAt:      &now,
				Result:          "success",
				OperatorType:    "system",
			}
			escalateLog := SubscriptionReconciliationModel.RenewalExecutionLog{
				ID:              utils.NewUUID(),
				SubscriptionRef: candidate.SubscriptionRef,
				ActionType:      "escalate",
				AttemptNo:       candidateResult.retryFailed,
				ExecutedAt:      &now,
				Result:          "success",
				OperatorType:    "system",
			}
			logs = append(logs, notifyLog, escalateLog)
			s.publisher.Emit(ctx, tenantUUID, SubscriptionReconciliationObs.EventRenewalEscalated, map[string]any{
				"subscription_ref": candidate.SubscriptionRef,
				"failed_attempts":  candidateResult.retryFailed,
			})
		}
	}
	if !dryRun {
		if err := s.executionRepo.CreateBatch(ctx, logs); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"retryTotal":     len(candidates),
		"retrySucceeded": retrySucceeded,
		"retryFailed":    retryFailed,
		"escalated":      escalated,
	}, nil
}
