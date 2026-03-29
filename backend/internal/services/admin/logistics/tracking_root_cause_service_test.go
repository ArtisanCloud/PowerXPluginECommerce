package logistics

import (
	"context"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTrackingRootCauseService_AttributionAccuracy(t *testing.T) {
	db := setupBillingDB(t, "logistics_root_cause_accuracy")
	ensureTrackingRootCauseTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-rc-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:              "wb-rc-1",
		TenantUUID:      "tenant-rc-a",
		OrderID:         "order-rc-1",
		CarrierID:       "carrier-rc-a",
		ServiceCode:     "std",
		WaybillNo:       "WB-RC-1",
		Status:          "delay",
		FeeAmount:       10,
		ActualFeeAmount: 10,
		Metadata:        []byte(`{"warehouse_id":"wh-a","destination_zone":"CN-EAST"}`),
		CreatedAt:       now.Add(-6 * time.Hour),
		UpdatedAt:       now.Add(-6 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingEvent{
		ID:         "evt-rc-1",
		TenantUUID: "tenant-rc-a",
		WaybillID:  "wb-rc-1",
		WaybillNo:  "WB-RC-1",
		EventID:    "event-1",
		Status:     "delay",
		Source:     "provider",
		CreatedAt:  now.Add(-5 * time.Hour),
		UpdatedAt:  now.Add(-5 * time.Hour),
	}).Error)

	svc := NewTrackingRootCauseService(&app.Deps{DB: db})
	rows, err := svc.Analyze(ctx, "tenant-rc-a", AnalyzeTrackingRootCauseRequest{WindowHours: 24, Limit: 50})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "delay", rows[0].AnomalyType)
	require.Equal(t, "carrier", rows[0].OwnerType)
	require.Equal(t, "expedite_route", rows[0].SuggestedAction)

	summary, err := svc.Summary(ctx, "tenant-rc-a", TrackingRootCauseQuery{Limit: 50})
	require.NoError(t, err)
	require.Equal(t, 1, summary.TotalCases)
	require.Equal(t, 1, summary.OpenCases)
	require.Equal(t, 0, summary.ResolvedCases)
}

func TestTrackingRootCauseService_ActionClosureIdempotent(t *testing.T) {
	db := setupBillingDB(t, "logistics_root_cause_action")
	ensureTrackingRootCauseTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-rc-b")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingRootCause{
		ID:              "rc-case-1",
		TenantUUID:      "tenant-rc-b",
		WaybillID:       "wb-rc-2",
		WaybillNo:       "WB-RC-2",
		CarrierID:       "carrier-rc-b",
		AnomalyType:     "tracking_exception",
		OwnerType:       "warehouse",
		Severity:        "high",
		SuggestedAction: "manual_review",
		Status:          "open",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error)

	svc := NewTrackingRootCauseService(&app.Deps{DB: db})
	first, err := svc.Handle(ctx, "tenant-rc-b", HandleTrackingRootCauseRequest{
		RootCauseID: "rc-case-1",
		Action:      "manual_review",
		Status:      "resolved",
		OperatorID:  "ops-1",
		ResultNote:  "fixed",
	})
	require.NoError(t, err)
	require.Equal(t, "resolved", first.Status)
	require.Equal(t, "manual_review", first.SelectedAction)
	require.NotNil(t, first.HandledAt)

	second, err := svc.Handle(ctx, "tenant-rc-b", HandleTrackingRootCauseRequest{
		RootCauseID: "rc-case-1",
		Action:      "manual_review",
		Status:      "resolved",
		OperatorID:  "ops-1",
	})
	require.NoError(t, err)
	require.Equal(t, "resolved", second.Status)
	require.Equal(t, "manual_review", second.SelectedAction)
}

func TestTrackingRootCauseService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_root_cause_tenant")
	ensureTrackingRootCauseTables(t, db)
	svc := NewTrackingRootCauseService(&app.Deps{DB: db})
	now := time.Now().UTC()

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-rc-x")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-rc-y")

	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.TrackingRootCause{
		ID:              "rc-tenant-a",
		TenantUUID:      "tenant-rc-x",
		WaybillID:       "wb-tenant-a",
		WaybillNo:       "WB-TENANT-A",
		CarrierID:       "carrier-a",
		AnomalyType:     "timeout",
		OwnerType:       "carrier",
		Severity:        "high",
		SuggestedAction: "priority_dispatch",
		Status:          "open",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error)

	rowsB, err := svc.List(ctxB, "tenant-rc-y", TrackingRootCauseQuery{Limit: 20})
	require.NoError(t, err)
	require.Len(t, rowsB, 0)

	_, err = svc.Handle(ctxB, "tenant-rc-y", HandleTrackingRootCauseRequest{
		RootCauseID: "rc-tenant-a",
		Action:      "manual_review",
		Status:      "resolved",
	})
	require.Error(t, err)
}

func ensureTrackingRootCauseTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_tracking_root_causes (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		anomaly_type TEXT NOT NULL,
		owner_type TEXT NOT NULL DEFAULT 'unknown',
		severity TEXT NOT NULL DEFAULT 'medium',
		suggested_action TEXT NOT NULL DEFAULT 'manual_review',
		selected_action TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		evidence_chain JSON,
		result_note TEXT,
		handled_by TEXT,
		detected_at DATETIME,
		handled_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_root_cause_waybill_type ON logistics_tracking_root_causes(tenant_uuid, waybill_id, anomaly_type)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_logistics_root_cause_scope ON logistics_tracking_root_causes(tenant_uuid, carrier_id, warehouse_id, destination_zone)`).Error)
}
