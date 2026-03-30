package logistics

import (
	"context"
	"strings"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFulfillmentFinanceRiskService_ScoringBoundaries(t *testing.T) {
	db := setupBillingDB(t, "logistics_finance_risk_scoring")
	ensureFinanceRiskTables(t, db)
	svc := NewFulfillmentFinanceRiskService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-fin-risk")

	require.NoError(t, db.WithContext(ctx).Exec(`
		INSERT INTO logistics_waybills (id, tenant_uuid, order_id, carrier_id, service_code, waybill_no, status, fee_amount, actual_fee_amount, fee_diff_amount, billing_status, created_at, updated_at)
		VALUES
			('wb-low', 'tenant-fin-risk', 'order-low', 'carrier-a', 'std', 'WB-LOW', 'created', 100, 102, 2, 'settled', datetime('now'), datetime('now')),
			('wb-med', 'tenant-fin-risk', 'order-med', 'carrier-a', 'std', 'WB-MED', 'created', 100, 140, 40, 'pending', datetime('now'), datetime('now')),
			('wb-high', 'tenant-fin-risk', 'order-high', 'carrier-a', 'std', 'WB-HIGH', 'created', 100, 400, 300, 'pending', datetime('now'), datetime('now'))
	`).Error)

	rows, err := svc.Evaluate(ctx, "tenant-fin-risk", EvaluateFinanceRiskRequest{CarrierID: "carrier-a", Threshold: 70})
	require.NoError(t, err)
	require.Len(t, rows, 3)

	levels := map[string]string{}
	for _, row := range rows {
		levels[row.WaybillNo] = row.RiskLevel
	}
	require.Equal(t, "low", levels["WB-LOW"])
	require.Equal(t, "medium", levels["WB-MED"])
	require.Equal(t, "high", levels["WB-HIGH"])
}

func TestFulfillmentFinanceRiskService_ExecuteActionIdempotent(t *testing.T) {
	db := setupBillingDB(t, "logistics_finance_risk_idempotent")
	ensureFinanceRiskTables(t, db)
	svc := NewFulfillmentFinanceRiskService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-fin-risk")

	require.NoError(t, db.WithContext(ctx).Exec(`
		INSERT INTO logistics_waybills (id, tenant_uuid, order_id, carrier_id, service_code, waybill_no, status, fee_amount, actual_fee_amount, fee_diff_amount, billing_status, created_at, updated_at)
		VALUES ('wb-1', 'tenant-fin-risk', 'order-1', 'carrier-a', 'std', 'WB-1', 'created', 100, 180, 80, 'pending', datetime('now'), datetime('now'))
	`).Error)

	rows, err := svc.Evaluate(ctx, "tenant-fin-risk", EvaluateFinanceRiskRequest{CarrierID: "carrier-a", Threshold: 60})
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	riskID := rows[0].ID

	_, status1, err := svc.ExecuteAction(ctx, "tenant-fin-risk", riskID, ExecuteFinanceRiskActionRequest{
		Action:     "manual_review",
		OperatorID: "admin",
		Note:       "first",
		RequestKey: "req-1",
	})
	require.NoError(t, err)
	require.Equal(t, "executed", status1)

	_, status2, err := svc.ExecuteAction(ctx, "tenant-fin-risk", riskID, ExecuteFinanceRiskActionRequest{
		Action:     "manual_review",
		OperatorID: "admin",
		Note:       "second",
		RequestKey: "req-1",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", status2)

	audits, err := svc.ListAudits(ctx, "tenant-fin-risk", riskID, 10)
	require.NoError(t, err)
	require.Len(t, audits, 1)
}

func TestFulfillmentFinanceRiskService_AuditCompleteness(t *testing.T) {
	db := setupBillingDB(t, "logistics_finance_risk_audit")
	ensureFinanceRiskTables(t, db)
	svc := NewFulfillmentFinanceRiskService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-fin-risk")

	require.NoError(t, db.WithContext(ctx).Exec(`
		INSERT INTO logistics_waybills (id, tenant_uuid, order_id, carrier_id, service_code, waybill_no, status, fee_amount, actual_fee_amount, fee_diff_amount, billing_status, created_at, updated_at)
		VALUES ('wb-2', 'tenant-fin-risk', 'order-2', 'carrier-b', 'std', 'WB-2', 'created', 100, 210, 110, 'pending', datetime('now'), datetime('now'))
	`).Error)

	rows, err := svc.Evaluate(ctx, "tenant-fin-risk", EvaluateFinanceRiskRequest{CarrierID: "carrier-b", Threshold: 65})
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	riskID := rows[0].ID

	updated, status, err := svc.ExecuteAction(ctx, "tenant-fin-risk", riskID, ExecuteFinanceRiskActionRequest{
		Action:     "freeze_settlement",
		OperatorID: "finance-admin",
		Note:       "high exposure",
		RequestKey: "req-audit-1",
	})
	require.NoError(t, err)
	require.Equal(t, "executed", status)
	require.Equal(t, "freeze_settlement", updated.LastAction)
	require.Equal(t, "finance-admin", updated.LastActionBy)
	require.NotNil(t, updated.LastActionAt)
	require.GreaterOrEqual(t, updated.ActionCount, 1)

	audits, err := svc.ListAudits(ctx, "tenant-fin-risk", riskID, 10)
	require.NoError(t, err)
	require.Len(t, audits, 1)
	audit := audits[0]
	require.Equal(t, "freeze_settlement", audit.Action)
	require.Equal(t, "finance-admin", audit.OperatorID)
	require.Equal(t, "high exposure", audit.Note)
	require.Equal(t, "req-audit-1", audit.RequestKey)
	require.NotZero(t, audit.CreatedAt)
	require.True(t, strings.Contains(string(audit.Payload), "stop_loss_action"))
	require.True(t, strings.Contains(string(audit.Payload), "execute_operator_note"))
}

func ensureFinanceRiskTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_finance_risks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		carrier_id TEXT,
		billing_case_id TEXT,
		payout_risk_score NUMERIC NOT NULL DEFAULT 0,
		chargeback_risk_score NUMERIC NOT NULL DEFAULT 0,
		composite_risk_score NUMERIC NOT NULL DEFAULT 0,
		risk_level TEXT NOT NULL DEFAULT 'low',
		threshold_value NUMERIC NOT NULL DEFAULT 70,
		stop_loss_action TEXT NOT NULL DEFAULT 'observe',
		status TEXT NOT NULL DEFAULT 'open',
		suggestion TEXT,
		risk_factors JSON,
		last_action TEXT,
		last_action_by TEXT,
		last_action_at DATETIME,
		action_count INTEGER NOT NULL DEFAULT 0,
		resolved_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_finance_risk_waybill ON logistics_finance_risks(tenant_uuid, waybill_id)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_finance_risk_audits (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		risk_id TEXT NOT NULL,
		request_key TEXT NOT NULL,
		action TEXT NOT NULL,
		operator_id TEXT,
		note TEXT,
		payload JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_finance_risk_audit_req ON logistics_finance_risk_audits(tenant_uuid, risk_id, request_key)`).Error)
}
