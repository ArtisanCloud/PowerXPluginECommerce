package fulfillment

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestWarehouseBridgeService_CreateExecuteRollback(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_warehouse_bridge")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	svc := NewWarehouseBridgeService(&app.Deps{DB: db})
	ctx := context.Background()
	tenantUUID := "tenant-a"

	task, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-wh-1",
		WarehouseID: "wh-a",
	})
	require.NoError(t, err)

	outbound, err := svc.CreateOutbound(ctx, tenantUUID, CreateOutboundRequest{
		TaskID: task.ID,
		Items: []PickLineItem{
			{SKU: "SKU-1", Qty: 2},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "reserved", outbound.Status)

	result, err := svc.ExecuteOutbound(ctx, tenantUUID, outbound.ID, ExecuteOutboundRequest{
		OperatorID: "admin",
		PackageNo:  1,
	})
	require.NoError(t, err)
	require.Equal(t, "shipped", result.Outbound.Status)
	require.Equal(t, "completed", result.Task.Status)

	rollback, err := svc.RollbackOutbound(ctx, tenantUUID, outbound.ID, "manual verify")
	require.NoError(t, err)
	require.Equal(t, "rollback", rollback.Outbound.Status)
	require.Equal(t, "exception", rollback.Task.Status)
}

func TestWarehouseBridgeService_TenantIsolation(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_warehouse_tenant")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	svc := NewWarehouseBridgeService(&app.Deps{DB: db})
	ctx := context.Background()

	task, err := taskSvc.Create(ctx, "tenant-a", CreateTaskRequest{
		OrderID:     "order-isolation",
		WarehouseID: "wh-a",
	})
	require.NoError(t, err)
	outbound, err := svc.CreateOutbound(ctx, "tenant-a", CreateOutboundRequest{
		TaskID: task.ID,
	})
	require.NoError(t, err)

	tenantBCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	_, err = svc.ExecuteOutbound(tenantBCtx, "tenant-b", outbound.ID, ExecuteOutboundRequest{OperatorID: "admin"})
	require.Error(t, err)
}
