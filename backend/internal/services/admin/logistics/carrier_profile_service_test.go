package logistics

import (
	"context"
	"testing"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestCarrierProfileService_EvaluateStableScore(t *testing.T) {
	db := setupBillingDB(t, "logistics_carrier_profile_eval_stable")
	ensureCarrierProfileTables(t, db)
	svc := NewCarrierProfileService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-profile-a")

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-profile-a",
		TenantUUID: "tenant-profile-a",
		Name:       "Carrier A",
		Code:       "carrier-a",
		Type:       "self",
		Status:     "active",
		Config:     datatypes.JSON([]byte(`{"onTimeRate":95.2,"avgHours":26,"costIndex":1.08}`)),
	}).Error)

	firstRows, err := svc.Evaluate(ctx, "tenant-profile-a", EvaluateCarrierProfilesRequest{
		CarrierID:  "carrier-profile-a",
		OperatorID: "operator-a",
	})
	require.NoError(t, err)
	require.Len(t, firstRows, 1)

	secondRows, err := svc.Evaluate(ctx, "tenant-profile-a", EvaluateCarrierProfilesRequest{
		CarrierID:  "carrier-profile-a",
		OperatorID: "operator-a",
	})
	require.NoError(t, err)
	require.Len(t, secondRows, 1)

	require.Equal(t, firstRows[0].CompositeScore, secondRows[0].CompositeScore)
	require.Equal(t, firstRows[0].StabilityScore, secondRows[0].StabilityScore)
	require.Equal(t, firstRows[0].CostScore, secondRows[0].CostScore)
	require.Equal(t, firstRows[0].ServiceRating, secondRows[0].ServiceRating)
}

func TestCarrierProfileService_RetireConstraintAndRestore(t *testing.T) {
	db := setupBillingDB(t, "logistics_carrier_profile_retire")
	ensureCarrierProfileTables(t, db)
	svc := NewCarrierProfileService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-profile-b")

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-profile-b",
		TenantUUID: "tenant-profile-b",
		Name:       "Carrier B",
		Code:       "carrier-b",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.CarrierProfile{
		ID:             "profile-b",
		TenantUUID:     "tenant-profile-b",
		CarrierID:      "carrier-profile-b",
		StabilityScore: 95,
		CostScore:      88,
		CompositeScore: 92,
		ServiceRating:  "A",
		Status:         "active",
		ScoreTrend:     datatypes.JSON([]byte(`[92]`)),
	}).Error)

	_, err := svc.Retire(ctx, "tenant-profile-b", "profile-b", RetireCarrierProfileRequest{
		Reason:     "manual-review",
		OperatorID: "operator-b",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "without force")

	row, err := svc.Retire(ctx, "tenant-profile-b", "profile-b", RetireCarrierProfileRequest{
		Reason:     "manual-force-retire",
		Force:      true,
		OperatorID: "operator-b",
	})
	require.NoError(t, err)
	require.Equal(t, "retired", row.Status)
	require.Equal(t, "manual-force-retire", row.RetireReason)

	restored, err := svc.Restore(ctx, "tenant-profile-b", "profile-b", RestoreCarrierProfileRequest{
		OperatorID: "operator-b",
	})
	require.NoError(t, err)
	require.Equal(t, "active", restored.Status)
	require.Empty(t, restored.RetireReason)
}

func TestCarrierProfileService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_carrier_profile_tenant")
	ensureCarrierProfileTables(t, db)
	svc := NewCarrierProfileService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-profile-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-profile-b")

	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.Carrier{
		ID:         "carrier-profile-ta",
		TenantUUID: "tenant-profile-a",
		Name:       "Carrier Tenant A",
		Code:       "carrier-ta",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.CarrierProfile{
		ID:             "profile-ta",
		TenantUUID:     "tenant-profile-a",
		CarrierID:      "carrier-profile-ta",
		StabilityScore: 82,
		CostScore:      81,
		CompositeScore: 81.6,
		ServiceRating:  "B",
		Status:         "active",
	}).Error)

	rowsB, err := svc.List(ctxB, "tenant-profile-b", CarrierProfileQuery{Limit: 20})
	require.NoError(t, err)
	require.Len(t, rowsB, 0)

	_, err = svc.Retire(ctxB, "tenant-profile-b", "profile-ta", RetireCarrierProfileRequest{
		Reason: "cross-tenant",
		Force:  true,
	})
	require.Error(t, err)
}

func ensureCarrierProfileTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_carrier_profiles (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		stability_score NUMERIC NOT NULL DEFAULT 0,
		cost_score NUMERIC NOT NULL DEFAULT 0,
		service_rating TEXT NOT NULL DEFAULT 'B',
		composite_score NUMERIC NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'active',
		retire_reason TEXT,
		score_trend JSON,
		suggestion TEXT,
		confirmed_by TEXT,
		confirmed_at DATETIME,
		evaluated_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_carrier_profile_carrier ON logistics_carrier_profiles(tenant_uuid, carrier_id)`).Error)
}
