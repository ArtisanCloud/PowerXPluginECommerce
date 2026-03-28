package fulfillment

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

func TestTaskService_StatusFlow(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_task_service_flow")
	svc := NewTaskService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-1"
	task, err := svc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-1001",
		WarehouseID: "warehouse-a",
		AssignedTo:  "picker-a",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", task.Status)

	task, err = svc.Advance(ctx, tenantUUID, task.ID, AdvanceTaskRequest{
		Status:     "picking",
		OperatorID: "picker-a",
	})
	require.NoError(t, err)
	require.Equal(t, "picking", task.Status)

	task, err = svc.Advance(ctx, tenantUUID, task.ID, AdvanceTaskRequest{
		Status:     "packed",
		OperatorID: "packer-a",
	})
	require.NoError(t, err)
	require.Equal(t, "packed", task.Status)

	task, err = svc.Complete(ctx, tenantUUID, task.ID, AdvanceTaskRequest{
		OperatorID: "checker-a",
	})
	require.NoError(t, err)
	require.Equal(t, "completed", task.Status)

	var logCount int64
	require.NoError(t, db.WithContext(authx.ContextWithTenantUUID(ctx, tenantUUID)).
		Table(coremodels.S(coremodels.TableFulfillmentTaskLogs)).
		Where("task_id = ?", task.ID).
		Count(&logCount).Error)
	require.GreaterOrEqual(t, logCount, int64(4))
}

func setupFulfillmentServiceDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_tasks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		order_id TEXT NOT NULL,
		warehouse_id TEXT NOT NULL,
		status TEXT NOT NULL,
		assigned_to TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_task_logs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		task_id TEXT NOT NULL,
		action TEXT NOT NULL,
		operator_id TEXT,
		detail JSON,
		created_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_waves (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		warehouse_id TEXT NOT NULL,
		status TEXT NOT NULL,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_wave_task_links (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		wave_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		result TEXT NOT NULL,
		message TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_fulfillment_wave_task ON fulfillment_wave_task_links(tenant_uuid, wave_id, task_id)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_wave_strategies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		warehouse_id TEXT,
		carrier_code TEXT,
		time_window TEXT,
		priority_band TEXT,
		max_tasks_per_wave INTEGER NOT NULL DEFAULT 50,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		rules JSON NOT NULL,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_exceptions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		task_id TEXT NOT NULL,
		waybill_id TEXT,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		reason TEXT,
		first_action_at DATETIME,
		escalated_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_outbounds (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		task_id TEXT NOT NULL,
		order_id TEXT NOT NULL,
		warehouse_id TEXT NOT NULL,
		waybill_id TEXT,
		status TEXT NOT NULL,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_fulfillment_outbound_task ON fulfillment_outbounds(tenant_uuid, task_id)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_pick_items (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		outbound_id TEXT NOT NULL,
		sku TEXT NOT NULL,
		qty INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_fulfillment_pick_line ON fulfillment_pick_items(tenant_uuid, outbound_id, sku)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS fulfillment_pack_orders (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		outbound_id TEXT NOT NULL,
		package_no INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_fulfillment_pack_outbound ON fulfillment_pack_orders(tenant_uuid, outbound_id)`).Error)

	return db
}

func TestExceptionService_EscalateOverdue(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_exception_service_escalate")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	exSvc := NewExceptionService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-1"
	task, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-2001",
		WarehouseID: "warehouse-b",
	})
	require.NoError(t, err)

	ex, err := exSvc.Report(ctx, tenantUUID, ReportExceptionRequest{
		TaskID: task.ID,
		Type:   "delay",
		Reason: "分拣延迟",
	})
	require.NoError(t, err)
	require.Equal(t, "open", ex.Status)

	oldCreatedAt := time.Now().UTC().Add(-25 * time.Hour)
	require.NoError(t, db.WithContext(authx.ContextWithTenantUUID(ctx, tenantUUID)).
		Table(coremodels.S(coremodels.TableFulfillmentExceptions)).
		Where("id = ?", ex.ID).
		Updates(map[string]any{
			"created_at":      oldCreatedAt,
			"first_action_at": nil,
			"status":          "open",
			"escalated_at":    nil,
		}).Error)

	escalated, err := exSvc.EscalateOverdue(ctx, tenantUUID, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, escalated, 1)
	require.Equal(t, ex.ID, escalated[0].ID)
	require.Equal(t, "escalated", escalated[0].Status)
	require.NotNil(t, escalated[0].EscalatedAt)
}
