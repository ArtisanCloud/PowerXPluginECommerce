package logistics

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	logisticsmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRoutingService_PriorityRuleWins(t *testing.T) {
	db := setupRoutingDB(t, "logistics_routing_priority")
	seedRoutingCarriers(t, db)
	require.NoError(t, db.Create(&logisticsmodel.RoutingRule{
		ID:              "r1",
		TenantUUID:      "tenant-1",
		Name:            "rule-low",
		WarehouseID:     "WH-A",
		DestinationZone: "CN-EAST",
		CarrierID:       "carrier-a",
		ServiceCode:     "std",
		Priority:        100,
		Enabled:         true,
	}).Error)
	require.NoError(t, db.Create(&logisticsmodel.RoutingRule{
		ID:              "r2",
		TenantUUID:      "tenant-1",
		Name:            "rule-high",
		WarehouseID:     "WH-A",
		DestinationZone: "CN-EAST",
		CarrierID:       "carrier-b",
		ServiceCode:     "std",
		Priority:        200,
		Enabled:         true,
	}).Error)

	svc := NewRoutingService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	result, err := svc.Preview(ctx, "tenant-1", PreviewRoutingRequest{
		WarehouseID:     "WH-A",
		DestinationZone: "CN-EAST",
		ServiceCode:     "std",
		Weight:          2,
	})
	require.NoError(t, err)
	require.Equal(t, "carrier-b", result.CarrierID)
	require.Equal(t, "r2", result.MatchedRuleID)
	require.Equal(t, "rule", result.Strategy)
}

func TestRoutingService_PreferredFallback(t *testing.T) {
	db := setupRoutingDB(t, "logistics_routing_fallback")
	seedRoutingCarriers(t, db)
	svc := NewRoutingService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	result, err := svc.Preview(ctx, "tenant-1", PreviewRoutingRequest{
		WarehouseID:        "WH-X",
		DestinationZone:    "CN-NORTH",
		PreferredCarrierID: "carrier-a",
		ServiceCode:        "next_day",
	})
	require.NoError(t, err)
	require.Equal(t, "preferred", result.Strategy)
	require.Equal(t, "carrier-a", result.CarrierID)
	require.True(t, result.Fallback)
}

func TestRoutingService_TenantIsolation(t *testing.T) {
	db := setupRoutingDB(t, "logistics_routing_tenant")
	seedRoutingCarriers(t, db)
	require.NoError(t, db.Create(&logisticsmodel.RoutingRule{
		ID:              "tenant2-rule",
		TenantUUID:      "tenant-2",
		Name:            "tenant2-rule",
		WarehouseID:     "WH-A",
		DestinationZone: "CN-EAST",
		CarrierID:       "carrier-t2",
		ServiceCode:     "std",
		Priority:        999,
		Enabled:         true,
	}).Error)
	svc := NewRoutingService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	result, err := svc.Preview(ctx, "tenant-1", PreviewRoutingRequest{
		WarehouseID:     "WH-A",
		DestinationZone: "CN-EAST",
		ServiceCode:     "std",
	})
	require.NoError(t, err)
	require.NotEqual(t, "carrier-t2", result.CarrierID)
}

func setupRoutingDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_carriers (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		contact_name TEXT,
		contact_phone TEXT,
		capabilities JSON,
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_routing_rules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		warehouse_id TEXT,
		destination_zone TEXT,
		carrier_id TEXT NOT NULL,
		service_code TEXT,
		priority INTEGER NOT NULL DEFAULT 100,
		min_weight NUMERIC NOT NULL DEFAULT 0,
		max_weight NUMERIC NOT NULL DEFAULT 0,
		min_order_amount NUMERIC NOT NULL DEFAULT 0,
		max_order_amount NUMERIC NOT NULL DEFAULT 0,
		fallback BOOLEAN NOT NULL DEFAULT FALSE,
		enabled BOOLEAN NOT NULL DEFAULT TRUE,
		rule_config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_routing_rule_name ON logistics_routing_rules(tenant_uuid, name)`).Error)
	return db
}

func seedRoutingCarriers(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Create(&logisticsmodel.Carrier{
		ID:         "carrier-a",
		TenantUUID: "tenant-1",
		Name:       "Carrier-A",
		Code:       "A",
		Type:       "self",
		Status:     "active",
		Config:     datatypes.JSON([]byte(`{"on_time_rate":93.5}`)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
	require.NoError(t, db.Create(&logisticsmodel.Carrier{
		ID:         "carrier-b",
		TenantUUID: "tenant-1",
		Name:       "Carrier-B",
		Code:       "B",
		Type:       "self",
		Status:     "active",
		Config:     datatypes.JSON([]byte(`{"on_time_rate":96.8}`)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
	require.NoError(t, db.Create(&logisticsmodel.Carrier{
		ID:         "carrier-t2",
		TenantUUID: "tenant-2",
		Name:       "Carrier-T2",
		Code:       "T2",
		Type:       "self",
		Status:     "active",
		Config:     datatypes.JSON([]byte(`{"on_time_rate":99.9}`)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
}
