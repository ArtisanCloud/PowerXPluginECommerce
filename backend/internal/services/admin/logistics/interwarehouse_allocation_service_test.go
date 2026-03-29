package logistics

import (
	"context"
	"testing"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestInterwarehouseAllocationService_SuggestRankingStable(t *testing.T) {
	db := setupBillingDB(t, "logistics_interwarehouse_ranking")
	ensureInterwarehouseAllocationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-iw-a")
	seedInterwarehousePlan(t, db, "tenant-iw-a", "plan-src", "carrier-iw", "wh-src", "CN-EAST", 20, 6, 15)
	seedInterwarehousePlan(t, db, "tenant-iw-a", "plan-t1", "carrier-iw", "wh-t1", "CN-EAST", 60, 5, 15)
	seedInterwarehousePlan(t, db, "tenant-iw-a", "plan-t2", "carrier-iw", "wh-t2", "CN-EAST", 55, 8, 20)
	seedInterwarehousePlan(t, db, "tenant-iw-a", "plan-t3", "carrier-iw", "wh-t3", "CN-EAST", 30, 3, 10)

	svc := NewInterwarehouseAllocationService(&app.Deps{DB: db})
	items, err := svc.Suggest(ctx, "tenant-iw-a", SuggestInterwarehouseAllocationRequest{
		RequestKey:        "iw-rank-1",
		CarrierID:         "carrier-iw",
		SourceWarehouseID: "wh-src",
		DestinationZone:   "CN-EAST",
		RequiredQty:       8,
		OperatorID:        "ops-iw",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(items), 2)

	for i := 1; i < len(items); i++ {
		prev := items[i-1]
		curr := items[i]
		if prev.Score == curr.Score {
			if prev.TargetAvailable == curr.TargetAvailable {
				require.LessOrEqual(t, prev.TargetWarehouseID, curr.TargetWarehouseID)
				continue
			}
			require.GreaterOrEqual(t, prev.TargetAvailable, curr.TargetAvailable)
			continue
		}
		require.GreaterOrEqual(t, prev.Score, curr.Score)
	}

	reloaded, err := svc.Suggest(ctx, "tenant-iw-a", SuggestInterwarehouseAllocationRequest{
		RequestKey:        "iw-rank-1",
		CarrierID:         "carrier-iw",
		SourceWarehouseID: "wh-src",
		DestinationZone:   "CN-EAST",
		RequiredQty:       8,
	})
	require.NoError(t, err)
	require.Len(t, reloaded, len(items))
	for i := range items {
		require.Equal(t, items[i].ID, reloaded[i].ID)
	}
}

func TestInterwarehouseAllocationService_TransferConstraints(t *testing.T) {
	db := setupBillingDB(t, "logistics_interwarehouse_constraints")
	ensureInterwarehouseAllocationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-iw-b")
	seedInterwarehousePlan(t, db, "tenant-iw-b", "plan-src", "carrier-iw", "wh-src", "CN-NORTH", 18, 2, 17)
	seedInterwarehousePlan(t, db, "tenant-iw-b", "plan-t1", "carrier-iw", "wh-t1", "CN-NORTH", 25, 3, 10)
	seedInterwarehousePlan(t, db, "tenant-iw-b", "plan-t2", "carrier-iw", "wh-t2", "CN-NORTH", 16, 2, 14)
	seedInterwarehousePlan(t, db, "tenant-iw-b", "plan-zero", "carrier-iw", "wh-zero", "CN-NORTH", 10, 4, 6)

	svc := NewInterwarehouseAllocationService(&app.Deps{DB: db})
	items, err := svc.Suggest(ctx, "tenant-iw-b", SuggestInterwarehouseAllocationRequest{
		RequestKey:        "iw-constraint-1",
		CarrierID:         "carrier-iw",
		SourceWarehouseID: "wh-src",
		DestinationZone:   "CN-NORTH",
		RequiredQty:       5,
	})
	require.NoError(t, err)
	require.NotEmpty(t, items)

	sourceAvailable := 18 - 2 - 17
	shortage := 5 - sourceAvailable
	require.Greater(t, shortage, 0)

	for _, item := range items {
		require.NotEqual(t, "wh-src", item.TargetWarehouseID)
		require.Greater(t, item.TargetAvailable, 0)
		require.Greater(t, item.TransferQty, 0)
		require.LessOrEqual(t, item.TransferQty, shortage)
		require.LessOrEqual(t, item.TransferQty, item.TargetAvailable)
	}
}

func TestInterwarehouseAllocationService_ConfirmIdempotent(t *testing.T) {
	db := setupBillingDB(t, "logistics_interwarehouse_confirm")
	ensureInterwarehouseAllocationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-iw-c")
	seedInterwarehousePlan(t, db, "tenant-iw-c", "plan-src", "carrier-iw", "wh-src", "CN-WEST", 18, 1, 17)
	seedInterwarehousePlan(t, db, "tenant-iw-c", "plan-t1", "carrier-iw", "wh-t1", "CN-WEST", 30, 5, 10)
	seedInterwarehousePlan(t, db, "tenant-iw-c", "plan-t2", "carrier-iw", "wh-t2", "CN-WEST", 28, 4, 10)

	svc := NewInterwarehouseAllocationService(&app.Deps{DB: db})
	items, err := svc.Suggest(ctx, "tenant-iw-c", SuggestInterwarehouseAllocationRequest{
		RequestKey:        "iw-confirm-1",
		CarrierID:         "carrier-iw",
		SourceWarehouseID: "wh-src",
		DestinationZone:   "CN-WEST",
		RequiredQty:       6,
	})
	require.NoError(t, err)
	require.NotEmpty(t, items)
	selected := items[0]

	first, err := svc.Confirm(ctx, "tenant-iw-c", ConfirmInterwarehouseAllocationRequest{
		CandidateID: selected.ID,
		OperatorID:  "ops-1",
	})
	require.NoError(t, err)
	require.Equal(t, "confirmed", first.Selected.Status)
	require.NotNil(t, first.Selected.ConfirmedAt)

	var srcAfterFirst LogisticsModel.CapacityPlan
	var targetAfterFirst LogisticsModel.CapacityPlan
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-iw-c", "plan-src").First(&srcAfterFirst).Error)
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-iw-c", "plan-t1").First(&targetAfterFirst).Error)
	srcUsedFirst := srcAfterFirst.UsedCapacity
	targetUsedFirst := targetAfterFirst.UsedCapacity

	second, err := svc.Confirm(ctx, "tenant-iw-c", ConfirmInterwarehouseAllocationRequest{
		CandidateID: selected.ID,
		OperatorID:  "ops-1",
	})
	require.NoError(t, err)
	require.Equal(t, first.Selected.ID, second.Selected.ID)
	require.Equal(t, "confirmed", second.Selected.Status)

	var srcAfterSecond LogisticsModel.CapacityPlan
	var targetAfterSecond LogisticsModel.CapacityPlan
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-iw-c", "plan-src").First(&srcAfterSecond).Error)
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-iw-c", "plan-t1").First(&targetAfterSecond).Error)
	require.Equal(t, srcUsedFirst, srcAfterSecond.UsedCapacity)
	require.Equal(t, targetUsedFirst, targetAfterSecond.UsedCapacity)

	for _, row := range second.Items {
		if row.ID == first.Selected.ID {
			require.Equal(t, "confirmed", row.Status)
			continue
		}
		require.NotEqual(t, "suggested", row.Status)
	}
}

func ensureInterwarehouseAllocationTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	ensureAllocationTables(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_interwarehouse_allocations (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		waybill_id TEXT,
		order_id TEXT,
		carrier_id TEXT,
		source_warehouse_id TEXT NOT NULL,
		target_warehouse_id TEXT NOT NULL,
		destination_zone TEXT,
		transfer_qty INTEGER NOT NULL DEFAULT 0,
		source_available INTEGER NOT NULL DEFAULT 0,
		target_available INTEGER NOT NULL DEFAULT 0,
		transfer_cost NUMERIC NOT NULL DEFAULT 0,
		eta_impact_hours NUMERIC NOT NULL DEFAULT 0,
		score NUMERIC NOT NULL DEFAULT 0,
		strategy TEXT NOT NULL DEFAULT 'collaboration_first',
		status TEXT NOT NULL DEFAULT 'suggested',
		reason TEXT,
		confirmed_by TEXT,
		confirmed_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_interwarehouse_candidate ON logistics_interwarehouse_allocations(tenant_uuid, request_key, target_warehouse_id)`).Error)
}

func seedInterwarehousePlan(
	t *testing.T,
	db *gorm.DB,
	tenantUUID string,
	id string,
	carrierID string,
	warehouseID string,
	destinationZone string,
	dailyCapacity int,
	reservedCapacity int,
	usedCapacity int,
) {
	t.Helper()
	require.NoError(t, db.Create(&LogisticsModel.CapacityPlan{
		ID:               id,
		TenantUUID:       tenantUUID,
		Name:             id,
		CarrierID:        carrierID,
		WarehouseID:      warehouseID,
		DestinationZone:  destinationZone,
		DailyCapacity:    dailyCapacity,
		ReservedCapacity: reservedCapacity,
		UsedCapacity:     usedCapacity,
		Status:           "active",
	}).Error)
}
