package subscription_reconciliation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHandlerContract_BatchDeltaTaskAdjustFlow(t *testing.T) {
	tenantUUID := "5a8180e8-d9f8-4de1-bf75-01fa68b46adb"
	r, db := setupContractRouter(t)

	createResp := performJSONRequest(t, r, http.MethodPost, "/admin/subscription-reconciliation/batches?tenant_uuid="+tenantUUID, map[string]any{
		"billingCycle": "2026-03-31",
		"runType":      "daily",
	})
	require.Equal(t, http.StatusOK, createResp.Code)
	batchID := nestedString(t, createResp.Body.Bytes(), "data", "id")
	require.NotEmpty(t, batchID)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas (
		id, tenant_uuid, batch_id, subscription_ref, bill_ref, payment_ref,
		delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor,
		reason_code, status, delta_fingerprint, detected_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"delta-contract-1", tenantUUID, batchID, "sub-contract-1", "bill-contract-1", "pay-contract-1",
		"missing_payment", "high", 100, 0, 100,
		"initial", "open", "fp-contract-1", now, now, now,
	).Error)

	listBatchResp := performJSONRequest(t, r, http.MethodGet, "/admin/subscription-reconciliation/batches?tenant_uuid="+tenantUUID+"&billingCycle=2026-03-31", nil)
	require.Equal(t, http.StatusOK, listBatchResp.Code)
	require.Len(t, nestedArray(t, listBatchResp.Body.Bytes(), "data", "items"), 1)

	listDeltaResp := performJSONRequest(t, r, http.MethodGet, fmt.Sprintf("/admin/subscription-reconciliation/batches/%s/deltas?tenant_uuid=%s", batchID, tenantUUID), nil)
	require.Equal(t, http.StatusOK, listDeltaResp.Code)
	deltas := nestedArray(t, listDeltaResp.Body.Bytes(), "data", "items")
	require.NotEmpty(t, deltas)
	deltaID, ok := deltas[0].(map[string]any)["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, deltaID)

	createTaskResp := performJSONRequest(t, r, http.MethodPost, fmt.Sprintf("/admin/subscription-reconciliation/deltas/%s/tasks?tenant_uuid=%s", deltaID, tenantUUID), map[string]any{
		"assignee": "ops-001",
		"slaLevel": "high",
		"note":     "contract create task",
	})
	require.Equal(t, http.StatusOK, createTaskResp.Code)
	require.NotEmpty(t, nestedString(t, createTaskResp.Body.Bytes(), "data", "id"))

	adjustResp := performJSONRequest(t, r, http.MethodPost, fmt.Sprintf("/admin/subscription-reconciliation/deltas/%s/adjust?tenant_uuid=%s", deltaID, tenantUUID), map[string]any{
		"expectedAmountMinor": 120,
		"actualAmountMinor":   100,
		"reasonCode":          "manual_adjust",
		"note":                "contract adjust",
	})
	require.Equal(t, http.StatusOK, adjustResp.Code)
	require.Equal(t, "processing", nestedString(t, adjustResp.Body.Bytes(), "data", "status"))
}

func setupContractRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:subscription_reconciliation_handler_contract?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_batches (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		billing_cycle TEXT NOT NULL,
		run_type TEXT NOT NULL,
		expected_amount_minor INTEGER NOT NULL DEFAULT 0,
		actual_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_count INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		created_by TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_sr_batches_cycle ON subscription_reconciliation_batches(tenant_uuid, billing_cycle, run_type)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_deltas (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		subscription_ref TEXT,
		bill_ref TEXT,
		payment_ref TEXT,
		delta_type TEXT NOT NULL,
		risk_level TEXT NOT NULL,
		expected_amount_minor INTEGER NOT NULL DEFAULT 0,
		actual_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_amount_minor INTEGER NOT NULL DEFAULT 0,
		reason_code TEXT,
		status TEXT NOT NULL,
		delta_fingerprint TEXT NOT NULL,
		detected_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_sr_deltas_batch ON subscription_reconciliation_deltas(tenant_uuid, batch_id, status)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_delta_tasks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		delta_id TEXT NOT NULL,
		delta_fingerprint TEXT NOT NULL,
		assignee TEXT,
		priority TEXT,
		sla_level TEXT NOT NULL,
		sla_deadline DATETIME,
		status TEXT NOT NULL,
		resolution TEXT,
		resolution_note TEXT,
		closed_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_sr_delta_tasks_delta ON subscription_reconciliation_delta_tasks(tenant_uuid, delta_id, status)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_governance_policies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		retry_windows JSON,
		escalation_threshold INTEGER NOT NULL DEFAULT 3,
		notify_channels JSON,
		version INTEGER NOT NULL DEFAULT 1,
		effective_from DATETIME,
		effective_to DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_execution_logs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		subscription_ref TEXT NOT NULL,
		action_type TEXT NOT NULL,
		attempt_no INTEGER NOT NULL DEFAULT 0,
		scheduled_at DATETIME,
		executed_at DATETIME,
		result TEXT NOT NULL,
		failure_reason TEXT,
		operator_type TEXT NOT NULL,
		operator_id TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)

	r := gin.New()
	RegisterRoutes(r.Group("/admin"), &app.Deps{DB: db})
	return r, db
}

func performJSONRequest(t *testing.T, r *gin.Engine, method, path string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func nestedString(t *testing.T, raw []byte, keys ...string) string {
	t.Helper()
	v := nestedValue(t, raw, keys...)
	s, ok := v.(string)
	require.True(t, ok)
	return s
}

func nestedArray(t *testing.T, raw []byte, keys ...string) []any {
	t.Helper()
	v := nestedValue(t, raw, keys...)
	items, ok := v.([]any)
	require.True(t, ok)
	return items
}

func nestedValue(t *testing.T, raw []byte, keys ...string) any {
	t.Helper()
	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))

	var current any = payload
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		require.True(t, ok)
		current, ok = obj[key]
		require.True(t, ok)
	}
	return current
}
