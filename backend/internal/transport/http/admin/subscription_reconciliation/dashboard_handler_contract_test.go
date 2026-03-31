package subscription_reconciliation

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDashboardHandlerContract_QueryAndExport(t *testing.T) {
	tenantUUID := "66120bce-f6ba-4e2e-ba8f-f4729f28029b"
	r, db := setupContractRouter(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_batches
		(id, tenant_uuid, billing_cycle, run_type, expected_amount_minor, actual_amount_minor, delta_amount_minor, delta_count, status, started_at, finished_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"batch-contract-dashboard", tenantUUID, "2026-03-31", "daily", 200, 100, 100, 1, "completed", now, now, now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas
		(id, tenant_uuid, batch_id, subscription_ref, delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor, reason_code, status, delta_fingerprint, detected_at, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"delta-contract-dashboard", tenantUUID, "batch-contract-dashboard", "sub-contract-dashboard", "missing_payment", "high", 200, 100, 100, "always_fail", "open", "fp-contract-dashboard", now, `{"channel":"app","plan":"pro","region":"CN"}`, now, now,
	).Error)

	resp := performJSONRequest(t, r, http.MethodGet, "/admin/subscription-reconciliation/dashboard?tenant_uuid="+tenantUUID+"&from=2026-03-30&to=2026-03-31&channel=app&plan=pro&region=CN&failureReason=always_fail", nil)
	require.Equal(t, http.StatusOK, resp.Code)
	require.NotNil(t, nestedValue(t, resp.Body.Bytes(), "data", "deltaRate"))

	exportResp := performJSONRequest(t, r, http.MethodGet, "/admin/subscription-reconciliation/dashboard/export?tenant_uuid="+tenantUUID, nil)
	require.Equal(t, http.StatusOK, exportResp.Code)
	require.Contains(t, exportResp.Header().Get("Content-Type"), "text/csv")
	require.Contains(t, exportResp.Body.String(), "metric,value")
	require.True(t, strings.Contains(exportResp.Body.String(), "deltaRate"))
}
