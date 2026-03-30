package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestAddressValidationService_CheckAndCache(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_address_validation")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	svc := NewAddressValidationService(&app.Deps{DB: db})

	row, err := svc.Check(ctx, "tenant-a", CheckAddressRequest{
		RequestKey: "req-1",
		WaybillNo:  "WB-OK",
		Address:    "北京市 朝阳区 测试地址",
	})
	require.NoError(t, err)
	require.False(t, row.Reachable)
	require.Equal(t, "high", row.RiskLevel)

	cached, err := svc.Check(ctx, "tenant-a", CheckAddressRequest{
		RequestKey: "req-1",
		WaybillNo:  "WB-OK",
		Address:    "北京市 朝阳区 测试地址",
	})
	require.NoError(t, err)
	require.Equal(t, row.ID, cached.ID)
}

func TestAddressValidationService_TenantIsolation(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_address_validation_tenant")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	svc := NewAddressValidationService(&app.Deps{DB: db})

	_, err := svc.Check(ctx, "tenant-a", CheckAddressRequest{
		RequestKey: "req-tenant-a",
		WaybillNo:  "WB-A",
		Address:    "上海市 徐汇区 漕溪路 100 号",
	})
	require.NoError(t, err)

	tenantBCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	rows, err := svc.List(tenantBCtx, "tenant-b", "", 20)
	require.NoError(t, err)
	require.Len(t, rows, 0)
}
