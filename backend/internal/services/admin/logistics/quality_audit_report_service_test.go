package logistics

import (
	"context"
	"strings"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestQualityAuditReportService_GenerateMetricsCorrectness(t *testing.T) {
	db := setupBillingDB(t, "logistics_quality_audit_metrics")
	ensureQualityAuditReportTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-qa-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:              "wb-qa-1",
		TenantUUID:      "tenant-qa-a",
		OrderID:         "order-qa-1",
		CarrierID:       "carrier-qa-a",
		ServiceCode:     "std",
		WaybillNo:       "WB-QA-1",
		Status:          "delivered",
		FeeAmount:       10,
		ActualFeeAmount: 12,
		Metadata:        []byte(`{"warehouse_id":"wh-qa-1","destination_zone":"CN-EAST"}`),
		CreatedAt:       now.Add(-48 * time.Hour),
		UpdatedAt:       now.Add(-24 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:              "wb-qa-2",
		TenantUUID:      "tenant-qa-a",
		OrderID:         "order-qa-2",
		CarrierID:       "carrier-qa-a",
		ServiceCode:     "std",
		WaybillNo:       "WB-QA-2",
		Status:          "delay",
		FeeAmount:       18,
		ActualFeeAmount: 0,
		Metadata:        []byte(`{"warehouse_id":"wh-qa-1","destination_zone":"CN-EAST"}`),
		CreatedAt:       now.Add(-36 * time.Hour),
		UpdatedAt:       now.Add(-4 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:              "wb-qa-3",
		TenantUUID:      "tenant-qa-a",
		OrderID:         "order-qa-3",
		CarrierID:       "carrier-qa-a",
		ServiceCode:     "std",
		WaybillNo:       "WB-QA-3",
		Status:          "in_transit",
		FeeAmount:       20,
		ActualFeeAmount: 22,
		Metadata:        []byte(`{"warehouse_id":"wh-qa-1","destination_zone":"CN-EAST"}`),
		CreatedAt:       now.Add(-96 * time.Hour),
		UpdatedAt:       now.Add(-2 * time.Hour),
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.CapacityForecast{
		ID:              "forecast-qa-1",
		TenantUUID:      "tenant-qa-a",
		CarrierID:       "carrier-qa-a",
		WarehouseID:     "wh-qa-1",
		DestinationZone: "CN-EAST",
		Status:          "suggested",
		CreatedAt:       now.Add(-6 * time.Hour),
		UpdatedAt:       now.Add(-6 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingRootCause{
		ID:              "root-qa-1",
		TenantUUID:      "tenant-qa-a",
		WaybillID:       "wb-qa-2",
		WaybillNo:       "WB-QA-2",
		CarrierID:       "carrier-qa-a",
		WarehouseID:     "wh-qa-1",
		DestinationZone: "CN-EAST",
		AnomalyType:     "delay",
		OwnerType:       "carrier",
		Severity:        "high",
		SuggestedAction: "expedite_route",
		Status:          "open",
		CreatedAt:       now.Add(-3 * time.Hour),
		UpdatedAt:       now.Add(-3 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.InterwarehouseAllocation{
		ID:                "iw-qa-1",
		TenantUUID:        "tenant-qa-a",
		RequestKey:        "qa-iw-1",
		CarrierID:         "carrier-qa-a",
		SourceWarehouseID: "wh-qa-src",
		TargetWarehouseID: "wh-qa-1",
		DestinationZone:   "CN-EAST",
		TransferQty:       2,
		TransferCost:      18,
		ETAImpactHours:    3,
		Score:             0.82,
		Status:            "confirmed",
		CreatedAt:         now.Add(-2 * time.Hour),
		UpdatedAt:         now.Add(-2 * time.Hour),
	}).Error)

	svc := NewQualityAuditReportService(&app.Deps{DB: db})
	report, err := svc.Generate(ctx, "tenant-qa-a", GenerateQualityAuditReportRequest{
		CarrierID:       "carrier-qa-a",
		WarehouseID:     "wh-qa-1",
		DestinationZone: "CN-EAST",
		WindowHours:     24 * 7,
		OperatorID:      "ops-qa",
	})
	require.NoError(t, err)
	require.Equal(t, 3, report.TotalWaybills)
	require.Equal(t, 1, report.DeliveredCount)
	require.Equal(t, 1, report.ExceptionCount)
	require.Equal(t, 1, report.TimeoutCount)
	require.Equal(t, 1, report.ForecastCount)
	require.Equal(t, 1, report.RootCauseCount)
	require.Equal(t, 1, report.InterwarehouseCount)
	require.NotEmpty(t, strings.TrimSpace(report.Conclusion))
	require.NotEmpty(t, strings.TrimSpace(string(report.ActionItems)))
}

func TestQualityAuditReportService_ExportStable(t *testing.T) {
	db := setupBillingDB(t, "logistics_quality_audit_export")
	ensureQualityAuditReportTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-qa-export")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:              "wb-qa-export",
		TenantUUID:      "tenant-qa-export",
		OrderID:         "order-qa-export",
		CarrierID:       "carrier-qa-export",
		ServiceCode:     "std",
		WaybillNo:       "WB-QA-EXPORT",
		Status:          "delivered",
		FeeAmount:       12,
		ActualFeeAmount: 13,
		Metadata:        []byte(`{"warehouse_id":"wh-e","destination_zone":"CN-SOUTH"}`),
		CreatedAt:       now.Add(-12 * time.Hour),
		UpdatedAt:       now.Add(-2 * time.Hour),
	}).Error)

	svc := NewQualityAuditReportService(&app.Deps{DB: db})
	first, err := svc.Generate(ctx, "tenant-qa-export", GenerateQualityAuditReportRequest{
		CarrierID:       "carrier-qa-export",
		WarehouseID:     "wh-e",
		DestinationZone: "CN-SOUTH",
		WindowHours:     24,
		OperatorID:      "ops-export",
	})
	require.NoError(t, err)
	require.NotEmpty(t, first.ID)

	csv1, err := svc.ExportCSV(ctx, "tenant-qa-export", QualityAuditReportQuery{
		CarrierID:       "carrier-qa-export",
		WarehouseID:     "wh-e",
		DestinationZone: "CN-SOUTH",
		Limit:           20,
	})
	require.NoError(t, err)
	require.Contains(t, csv1, "id,period_from,period_to,carrier_id,warehouse_id,destination_zone")
	require.Contains(t, csv1, first.ID)

	csv2, err := svc.ExportCSV(ctx, "tenant-qa-export", QualityAuditReportQuery{
		CarrierID:       "carrier-qa-export",
		WarehouseID:     "wh-e",
		DestinationZone: "CN-SOUTH",
		Limit:           20,
	})
	require.NoError(t, err)
	require.Equal(t, csv1, csv2)
}

func TestQualityAuditReportService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_quality_audit_tenant")
	ensureQualityAuditReportTables(t, db)
	svc := NewQualityAuditReportService(&app.Deps{DB: db})
	now := time.Now().UTC()

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-qa-x")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-qa-y")

	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.QualityAuditReport{
		ID:               "report-qa-x",
		TenantUUID:       "tenant-qa-x",
		ReportPeriodFrom: now.Add(-24 * time.Hour),
		ReportPeriodTo:   now,
		WindowHours:      24,
		Status:           "generated",
		Conclusion:       "ok",
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error)

	rows, err := svc.List(ctxB, "tenant-qa-y", QualityAuditReportQuery{Limit: 20})
	require.NoError(t, err)
	require.Len(t, rows, 0)

	_, err = svc.Get(ctxB, "tenant-qa-y", "report-qa-x")
	require.Error(t, err)
}

func ensureQualityAuditReportTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	ensureCapacityForecastTables(t, db)
	ensureTrackingRootCauseTables(t, db)
	ensureInterwarehouseAllocationTables(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_quality_audit_reports (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		report_period_from DATETIME NOT NULL,
		report_period_to DATETIME NOT NULL,
		window_hours INTEGER NOT NULL DEFAULT 24,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		total_waybills INTEGER NOT NULL DEFAULT 0,
		delivered_count INTEGER NOT NULL DEFAULT 0,
		exception_count INTEGER NOT NULL DEFAULT 0,
		timeout_count INTEGER NOT NULL DEFAULT 0,
		on_time_rate NUMERIC NOT NULL DEFAULT 0,
		delivery_success_rate NUMERIC NOT NULL DEFAULT 0,
		total_cost NUMERIC NOT NULL DEFAULT 0,
		avg_cost NUMERIC NOT NULL DEFAULT 0,
		forecast_count INTEGER NOT NULL DEFAULT 0,
		root_cause_count INTEGER NOT NULL DEFAULT 0,
		interwarehouse_count INTEGER NOT NULL DEFAULT 0,
		conclusion TEXT,
		action_items JSON,
		status TEXT NOT NULL DEFAULT 'generated',
		generated_by TEXT,
		generated_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_logistics_quality_audit_scope ON logistics_quality_audit_reports(tenant_uuid, carrier_id, warehouse_id, destination_zone)`).Error)
}
