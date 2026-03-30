package fulfillment

import (
	"context"
	"encoding/json"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestWaveService_BatchAdvancePartialFailure(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_wave_partial_failure")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	waveSvc := NewWaveService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-wave-1"
	task1, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-wave-1",
		WarehouseID: "wh-1",
		AssignedTo:  "picker-1",
	})
	require.NoError(t, err)
	task2, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-wave-2",
		WarehouseID: "wh-1",
		AssignedTo:  "picker-2",
	})
	require.NoError(t, err)

	wave, err := waveSvc.Create(ctx, tenantUUID, CreateWaveRequest{
		Name:        "morning-wave",
		WarehouseID: "wh-1",
		TaskIDs:     []string{task1.ID, task2.ID},
	})
	require.NoError(t, err)
	require.Len(t, wave.Tasks, 2)

	result, err := waveSvc.BatchAdvance(ctx, tenantUUID, wave.Wave.ID, BatchAdvanceWaveRequest{
		Status:      "picking",
		OperatorID:  "lead-a",
		FailTaskIDs: []string{task2.ID},
		Reason:      "stock lock failed",
	})
	require.NoError(t, err)
	require.Equal(t, "partially_failed", result.WaveStatus)
	require.Equal(t, 1, result.Success)
	require.Equal(t, 1, result.Failed)
}

func TestWaveService_ReassignTask(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_wave_reassign")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	waveSvc := NewWaveService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-wave-2"
	task, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-wave-3",
		WarehouseID: "wh-2",
		AssignedTo:  "picker-old",
	})
	require.NoError(t, err)
	wave, err := waveSvc.Create(ctx, tenantUUID, CreateWaveRequest{
		Name:        "noon-wave",
		WarehouseID: "wh-2",
		TaskIDs:     []string{task.ID},
	})
	require.NoError(t, err)

	updated, err := waveSvc.ReassignTask(ctx, tenantUUID, wave.Wave.ID, ReassignWaveTaskRequest{
		TaskID:     task.ID,
		AssignedTo: "picker-new",
		OperatorID: "lead-b",
	})
	require.NoError(t, err)
	require.Equal(t, "picker-new", updated.AssignedTo)
}

func TestWaveService_BatchAdvanceWritesAuditFields(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_wave_audit")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	waveSvc := NewWaveService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-wave-3"
	task, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-wave-4",
		WarehouseID: "wh-3",
		AssignedTo:  "picker-audit",
	})
	require.NoError(t, err)
	wave, err := waveSvc.Create(ctx, tenantUUID, CreateWaveRequest{
		Name:        "audit-wave",
		WarehouseID: "wh-3",
		TaskIDs:     []string{task.ID},
	})
	require.NoError(t, err)

	_, err = waveSvc.BatchAdvance(ctx, tenantUUID, wave.Wave.ID, BatchAdvanceWaveRequest{
		Status:     "picking",
		OperatorID: "lead-audit",
	})
	require.NoError(t, err)

	var logs []struct {
		Action string
		Detail string
	}
	require.NoError(t, db.WithContext(authx.ContextWithTenantUUID(ctx, tenantUUID)).
		Table(coremodels.S(coremodels.TableFulfillmentTaskLogs)).
		Select("action, detail").
		Where("task_id = ?", task.ID).
		Order("created_at DESC").
		Find(&logs).Error)
	require.NotEmpty(t, logs)

	found := false
	for _, item := range logs {
		if item.Action != "task.wave.batch.picking" {
			continue
		}
		var payload map[string]any
		_ = json.Unmarshal([]byte(item.Detail), &payload)
		if payload["wave_id"] == wave.Wave.ID && payload["result"] == "success" {
			found = true
			break
		}
	}
	require.True(t, found, "expected wave batch audit detail with wave_id/result")
}
