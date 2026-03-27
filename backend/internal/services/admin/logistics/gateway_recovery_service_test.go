package logistics

import (
	"context"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestGatewayRecoveryService_IngestAndCompensate(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "gateway_recovery_ingest")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)

	now := time.Now().UTC()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-failed-1",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-a",
		Provider:      "self",
		Status:        "failed",
		WaybillStatus: "in_transit",
		TotalWaybills: 1,
		SuccessCount:  0,
		FailedCount:   1,
		LastError:     "gateway status=401 body=missing authorization",
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error)

	svc := NewGatewayRecoveryService(&app.Deps{DB: db})
	created, err := svc.IngestFromFailedJobs(ctx, "tenant-a", 24)
	require.NoError(t, err)
	require.GreaterOrEqual(t, created, 1)

	rows, err := svc.List(ctx, "tenant-a", GatewayFailureQuery{Status: "pending", Limit: 20})
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	require.Equal(t, "auth", rows[0].ErrorClass)

	// bind waybill for compensation
	rows[0].WaybillID = "wb-ok"
	rows[0].WaybillNo = "WB-OK"
	rows[0].CircuitOpenTill = nil
	require.NoError(t, db.WithContext(ctx).Save(&rows[0]).Error)

	row, err := svc.Compensate(ctx, "tenant-a", rows[0].ID)
	require.NoError(t, err)
	require.Equal(t, "recovered", row.Status)
}

func TestGatewayRecoveryService_ClassifyError(t *testing.T) {
	class, code := classifyGatewayError("gateway status=502 body=bad gateway")
	require.Equal(t, "server_5xx", class)
	require.Equal(t, "E_5XX", code)

	class, code = classifyGatewayError("request timeout exceeded")
	require.Equal(t, "timeout", class)
	require.Equal(t, "E_TIMEOUT", code)
}
