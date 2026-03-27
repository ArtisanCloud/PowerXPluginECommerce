package logistics

import (
	"context"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestTrackingSyncSchedulerService_RunDueWithDedupe(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "tracking_sync_scheduler_due")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)

	svc := NewTrackingSyncSchedulerService(&app.Deps{DB: db})
	enabled := true
	schedule, err := svc.Upsert(ctx, "tenant-a", UpsertTrackingSyncScheduleRequest{
		Name:            "每15分钟同步",
		CronExpr:        "*/15 * * * *",
		WaybillStatus:   "in_transit",
		Enabled:         &enabled,
		DedupeWindowSec: 600,
		BatchLimit:      20,
		EventLimit:      20,
	})
	require.NoError(t, err)
	require.NotNil(t, schedule)
	past := time.Now().UTC().Add(-1 * time.Minute)
	schedule.NextTriggerAt = &past
	require.NoError(t, db.WithContext(ctx).Save(schedule).Error)

	triggered1, err := svc.RunDue(ctx, "tenant-a", 10)
	require.NoError(t, err)
	require.Equal(t, 1, triggered1)

	triggered2, err := svc.RunDue(ctx, "tenant-a", 10)
	require.NoError(t, err)
	require.Equal(t, 0, triggered2)

	updated, err := svc.Toggle(ctx, "tenant-a", schedule.ID, false)
	require.NoError(t, err)
	require.False(t, updated.Enabled)
}

func TestTrackingSyncSchedulerService_TriggerNow(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "tracking_sync_scheduler_trigger")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)

	svc := NewTrackingSyncSchedulerService(&app.Deps{DB: db})
	enabled := true
	schedule, err := svc.Upsert(ctx, "tenant-a", UpsertTrackingSyncScheduleRequest{
		Name:          "立即执行计划",
		CronExpr:      "*/5 * * * *",
		Enabled:       &enabled,
		WaybillStatus: "in_transit",
	})
	require.NoError(t, err)

	job, err := svc.TriggerNow(ctx, "tenant-a", schedule.ID)
	require.NoError(t, err)
	require.NotNil(t, job)
	require.True(t, job.SuccessCount+job.FailedCount >= 1)

	rows, err := svc.List(ctx, "tenant-a", nil, 20)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.True(t, rows[0].CreatedAt.Before(time.Now().Add(1*time.Second)))
}
