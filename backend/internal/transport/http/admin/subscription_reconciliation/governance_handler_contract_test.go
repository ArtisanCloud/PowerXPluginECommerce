package subscription_reconciliation

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGovernanceHandlerContract_RunGovernance(t *testing.T) {
	tenantUUID := "37c701e2-2e40-47e5-80a0-f5fe55b7c113"
	r, db := setupContractRouter(t)

	createResp := performJSONRequest(t, r, http.MethodPost, "/admin/subscription-reconciliation/batches?tenant_uuid="+tenantUUID, map[string]any{
		"billingCycle": "2026-03-31",
		"runType":      "daily",
	})
	require.Equal(t, http.StatusOK, createResp.Code)
	batchID := nestedString(t, createResp.Body.Bytes(), "data", "id")

	now := time.Now().UTC().Format(time.RFC3339Nano)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas (
		id, tenant_uuid, batch_id, subscription_ref, bill_ref, payment_ref,
		delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor,
		reason_code, status, delta_fingerprint, detected_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"delta-governance-1", tenantUUID, batchID, "sub-governance-1", "bill-governance-1", "pay-governance-1",
		"missing_payment", "high", 100, 0, 100,
		"always_fail", "open", "fp-governance-1", now, now, now,
	).Error)

	resp := performJSONRequest(t, r, http.MethodPost, "/admin/subscription-reconciliation/governance/run?tenant_uuid="+tenantUUID, map[string]any{
		"billingCycle": "2026-03-31",
	})
	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, float64(1), nestedValue(t, resp.Body.Bytes(), "data", "retryTotal"))
	require.Equal(t, float64(1), nestedValue(t, resp.Body.Bytes(), "data", "escalated"))

	var logs int64
	require.NoError(t, db.Table("subscription_reconciliation_execution_logs").Where("tenant_uuid = ?", tenantUUID).Count(&logs).Error)
	require.Greater(t, logs, int64(0), fmt.Sprintf("expected execution logs > 0, got %d", logs))
}
