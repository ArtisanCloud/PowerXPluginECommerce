package reverse

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReverseWaybillService_StatusFlowAndWarehouseResult(t *testing.T) {
	db := setupReverseServiceDB(t, "reverse_waybill_service_flow")
	svc := NewWaybillService(&app.Deps{DB: db})
	ctx := context.Background()
	tenantUUID := "tenant-1"

	waybill, err := svc.Create(ctx, tenantUUID, CreateWaybillRequest{
		OrderID:     "order-3001",
		AfterSaleID: "after-sale-3001",
	})
	require.NoError(t, err)
	require.Equal(t, "created", waybill.Status)

	_, _, err = svc.RecordWarehouseResult(ctx, tenantUUID, waybill.ID, RecordWarehouseResultRequest{
		Result:      "resellable",
		Disposition: "restock",
	})
	require.Error(t, err)

	_, err = svc.AppendTracking(ctx, tenantUUID, waybill.ID, AppendTrackingRequest{
		Status:      "in_transit",
		Description: "已揽收",
		OccurredAt:  ptrTime(time.Now().UTC()),
	})
	require.NoError(t, err)

	_, err = svc.AppendTracking(ctx, tenantUUID, waybill.ID, AppendTrackingRequest{
		Status:      "received",
		Description: "仓库签收",
		OccurredAt:  ptrTime(time.Now().UTC()),
	})
	require.NoError(t, err)

	result, updated, err := svc.RecordWarehouseResult(ctx, tenantUUID, waybill.ID, RecordWarehouseResultRequest{
		Result:      "damaged",
		Disposition: "compensate",
		OperatorID:  "qa-1",
		Notes:       "外包装破损",
	})
	require.NoError(t, err)
	require.Equal(t, "damaged", result.Result)
	require.Equal(t, "closed", updated.Status)
	require.Equal(t, "compensate", updated.Disposition)

	detail, err := svc.Detail(ctx, tenantUUID, waybill.ID)
	require.NoError(t, err)
	require.Equal(t, "refund_or_replacement", detail.Compensation["recommendation"])
}

func setupReverseServiceDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS reverse_waybills (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		order_id TEXT NOT NULL,
		after_sale_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		status TEXT NOT NULL,
		inspection_result TEXT,
		disposition TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_reverse_waybill_no ON reverse_waybills(tenant_uuid, waybill_no)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS reverse_waybill_tracking_events (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		status TEXT NOT NULL,
		description TEXT,
		occurred_at DATETIME,
		payload JSON,
		created_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS reverse_waybill_warehouse_results (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		result TEXT NOT NULL,
		operator_id TEXT,
		notes TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS reverse_inspection_rules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 100,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		condition_json JSON NOT NULL,
		decision TEXT NOT NULL,
		recommendation TEXT NOT NULL,
		notes TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS reverse_waybill_inspections (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		rule_id TEXT,
		decision TEXT NOT NULL,
		recommendation TEXT NOT NULL,
		reason TEXT,
		attributes JSON NOT NULL,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_reverse_waybill_inspection ON reverse_waybill_inspections(tenant_uuid, waybill_id)`).Error)

	return db
}

func ptrTime(v time.Time) *time.Time {
	return &v
}

func TestReverseWaybillService_TransitionGuard(t *testing.T) {
	db := setupReverseServiceDB(t, "reverse_waybill_transition_guard")
	svc := NewWaybillService(&app.Deps{DB: db})

	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-2")
	waybill, err := svc.Create(ctx, "tenant-2", CreateWaybillRequest{
		OrderID:     "order-guard",
		AfterSaleID: "after-sale-guard",
	})
	require.NoError(t, err)

	_, err = svc.AppendTracking(ctx, "tenant-2", waybill.ID, AppendTrackingRequest{
		Status: "closed",
	})
	require.Error(t, err)
}

func TestInspectionService_PriorityAndIdempotency(t *testing.T) {
	db := setupReverseServiceDB(t, "reverse_inspection_priority_idempotency")
	waybillSvc := NewWaybillService(&app.Deps{DB: db})
	inspectionSvc := NewInspectionService(&app.Deps{DB: db})
	ctx := context.Background()
	tenantUUID := "tenant-inspection"

	waybill, err := waybillSvc.Create(ctx, tenantUUID, CreateWaybillRequest{
		OrderID:     "order-inspection",
		AfterSaleID: "after-sale-inspection",
		Metadata: map[string]any{
			"package_status": "opened",
		},
	})
	require.NoError(t, err)
	_, err = waybillSvc.AppendTracking(ctx, tenantUUID, waybill.ID, AppendTrackingRequest{Status: "in_transit"})
	require.NoError(t, err)
	_, err = waybillSvc.AppendTracking(ctx, tenantUUID, waybill.ID, AppendTrackingRequest{Status: "received"})
	require.NoError(t, err)

	_, err = inspectionSvc.CreateRule(ctx, tenantUUID, CreateInspectionRuleRequest{
		Name:           "开封走维修",
		Priority:       20,
		Condition:      map[string]any{"package_status": "opened"},
		Decision:       "repair",
		Recommendation: "repair",
	})
	require.NoError(t, err)
	_, err = inspectionSvc.CreateRule(ctx, tenantUUID, CreateInspectionRuleRequest{
		Name:           "默认可二销",
		Priority:       30,
		Condition:      map[string]any{"package_status": "good"},
		Decision:       "resellable",
		Recommendation: "restock",
	})
	require.NoError(t, err)

	result1, err := inspectionSvc.EvaluateWaybill(ctx, tenantUUID, waybill.ID, EvaluateInspectionRequest{
		Attributes: map[string]any{"package_status": "opened"},
	})
	require.NoError(t, err)
	require.Equal(t, "repair", result1.Decision)
	require.Equal(t, "created", result1.IdempotencyState)
	require.NotNil(t, result1.Rule)
	require.Equal(t, "开封走维修", result1.Rule.Name)

	result2, err := inspectionSvc.EvaluateWaybill(ctx, tenantUUID, waybill.ID, EvaluateInspectionRequest{
		Attributes: map[string]any{"package_status": "good"},
	})
	require.NoError(t, err)
	require.Equal(t, "repair", result2.Decision)
	require.Equal(t, "replayed", result2.IdempotencyState)
}

func TestInspectionService_TenantIsolation(t *testing.T) {
	db := setupReverseServiceDB(t, "reverse_inspection_tenant_isolation")
	inspectionSvc := NewInspectionService(&app.Deps{DB: db})

	_, err := inspectionSvc.CreateRule(context.Background(), "tenant-A", CreateInspectionRuleRequest{
		Name:           "tenant-A-rule",
		Priority:       10,
		Condition:      map[string]any{"package_status": "good"},
		Decision:       "resellable",
		Recommendation: "restock",
	})
	require.NoError(t, err)
	_, err = inspectionSvc.CreateRule(context.Background(), "tenant-B", CreateInspectionRuleRequest{
		Name:           "tenant-B-rule",
		Priority:       10,
		Condition:      map[string]any{"package_status": "good"},
		Decision:       "damaged",
		Recommendation: "compensate",
	})
	require.NoError(t, err)

	rulesA, err := inspectionSvc.ListRules(context.Background(), "tenant-A")
	require.NoError(t, err)
	require.Len(t, rulesA, 1)
	require.Equal(t, "tenant-A-rule", rulesA[0].Name)

	rulesB, err := inspectionSvc.ListRules(context.Background(), "tenant-B")
	require.NoError(t, err)
	require.Len(t, rulesB, 1)
	require.Equal(t, "tenant-B-rule", rulesB[0].Name)
}
