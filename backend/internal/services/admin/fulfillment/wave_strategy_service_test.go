package fulfillment

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestWaveStrategyService_PreviewStableGrouping(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_wave_strategy_stable")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	strategySvc := NewWaveStrategyService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-wave-strategy-1"
	task1, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-s-1",
		WarehouseID: "wh-s-1",
		Metadata: map[string]any{
			"carrier_code": "sf",
			"priority":     "high",
			"time_slot":    "morning",
		},
	})
	require.NoError(t, err)
	task2, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-s-2",
		WarehouseID: "wh-s-1",
		Metadata: map[string]any{
			"carrier_code": "sf",
			"priority":     "high",
			"time_slot":    "morning",
		},
	})
	require.NoError(t, err)

	strategy, err := strategySvc.Create(ctx, tenantUUID, CreateWaveStrategyRequest{
		Name:         "morning-sf-high",
		WarehouseID:  "wh-s-1",
		CarrierCode:  "sf",
		TimeWindow:   "morning",
		PriorityBand: "high",
	})
	require.NoError(t, err)

	preview1, err := strategySvc.Preview(ctx, tenantUUID, PreviewWaveStrategyRequest{StrategyID: strategy.ID})
	require.NoError(t, err)
	preview2, err := strategySvc.Preview(ctx, tenantUUID, PreviewWaveStrategyRequest{StrategyID: strategy.ID})
	require.NoError(t, err)

	require.Len(t, preview1.Groups, 1)
	require.Len(t, preview2.Groups, 1)
	require.Equal(t, preview1.Groups[0].TaskIDs, preview2.Groups[0].TaskIDs)
	require.ElementsMatch(t, []string{task1.ID, task2.ID}, preview1.Groups[0].TaskIDs)
}

func TestWaveStrategyService_PreviewSkipsExceptionTasks(t *testing.T) {
	db := setupFulfillmentServiceDB(t, "fulfillment_wave_strategy_exception_isolation")
	taskSvc := NewTaskService(&app.Deps{DB: db})
	strategySvc := NewWaveStrategyService(&app.Deps{DB: db})

	ctx := context.Background()
	tenantUUID := "tenant-wave-strategy-2"
	okTask, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-e-1",
		WarehouseID: "wh-e-1",
		Metadata: map[string]any{
			"carrier_code": "yt",
			"priority":     "normal",
			"time_slot":    "afternoon",
		},
	})
	require.NoError(t, err)
	exTask, err := taskSvc.Create(ctx, tenantUUID, CreateTaskRequest{
		OrderID:     "order-e-2",
		WarehouseID: "wh-e-1",
		Metadata: map[string]any{
			"carrier_code": "yt",
			"priority":     "normal",
			"time_slot":    "afternoon",
		},
	})
	require.NoError(t, err)
	_, err = taskSvc.Advance(ctx, tenantUUID, exTask.ID, AdvanceTaskRequest{Status: "exception"})
	require.NoError(t, err)

	strategy, err := strategySvc.Create(ctx, tenantUUID, CreateWaveStrategyRequest{
		Name:        "afternoon-yt",
		WarehouseID: "wh-e-1",
		CarrierCode: "yt",
		TimeWindow:  "afternoon",
	})
	require.NoError(t, err)

	preview, err := strategySvc.Preview(ctx, tenantUUID, PreviewWaveStrategyRequest{StrategyID: strategy.ID})
	require.NoError(t, err)
	require.Len(t, preview.Groups, 1)
	require.Equal(t, []string{okTask.ID}, preview.Groups[0].TaskIDs)
}
