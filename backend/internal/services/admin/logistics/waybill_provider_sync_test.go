package logistics

import (
	"context"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics/integrations"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestWaybillService_SyncTrackingFromProvider(t *testing.T) {
	db := setupWaybillPartialShippingDB(t, "logistics_provider_sync")
	ctx := withTenantContext(context.Background(), "tenant-1")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-sync-1",
		TenantUUID: "tenant-1",
		Name:       "顺丰",
		Code:       "sf",
		Type:       "sf",
		Status:     "active",
		Config:     datatypes.JSON([]byte(`{"provider":"sf"}`)),
	}).Error)

	svc := NewWaybillService(&app.Deps{DB: db})
	adapter := &stubSyncAdapter{events: []integrations.TrackingFetchEvent{
		{EventID: "ev-1", Status: "in_transit", Description: "已揽收", OccurredAt: &now},
		{EventID: "ev-2", Status: "delivered", Description: "已签收", OccurredAt: &now},
	}}
	svc.adapters["sf"] = adapter

	wb, _, err := svc.Create(ctx, "tenant-1", CreateWaybillRequest{
		OrderID:     "order-sync-1",
		CarrierID:   "carrier-sync-1",
		ServiceCode: "std",
		WaybillNo:   "WB-SYNC-1",
	})
	require.NoError(t, err)

	result, err := svc.SyncTrackingFromProvider(ctx, "tenant-1", wb.ID, 10)
	require.NoError(t, err)
	require.Equal(t, 2, result.TotalFetched)
	require.Equal(t, 2, result.Appended)
	require.Equal(t, 0, result.Replayed)
	require.Equal(t, "delivered", result.CurrentStatus)
	require.Equal(t, 1, adapter.fetchCalls)

	again, err := svc.SyncTrackingFromProvider(ctx, "tenant-1", wb.ID, 10)
	require.NoError(t, err)
	require.Equal(t, 0, again.Appended)
	require.Equal(t, 2, again.Replayed)
	require.Equal(t, "delivered", again.CurrentStatus)
}

type stubSyncAdapter struct {
	fetchCalls int
	events     []integrations.TrackingFetchEvent
}

func (s *stubSyncAdapter) Provider() string { return "sf" }

func (s *stubSyncAdapter) TestConnectivity(_ context.Context, _ *LogisticsModel.Carrier) error {
	return nil
}

func (s *stubSyncAdapter) CreateWaybill(_ context.Context, _ *LogisticsModel.Carrier, input integrations.WaybillCreateInput) (*integrations.WaybillCreateResult, error) {
	waybillNo := input.WaybillNo
	if waybillNo == "" {
		waybillNo = "WB-STUB"
	}
	return &integrations.WaybillCreateResult{WaybillNo: waybillNo, Status: "created"}, nil
}

func (s *stubSyncAdapter) FetchTracking(_ context.Context, _ *LogisticsModel.Carrier, _ integrations.TrackingFetchInput) ([]integrations.TrackingFetchEvent, error) {
	s.fetchCalls++
	return s.events, nil
}

func (s *stubSyncAdapter) NormalizeTrackingStatus(status string) string { return status }
