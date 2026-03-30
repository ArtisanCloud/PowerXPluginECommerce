package integrations

import (
	"context"
	"testing"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestSelfAdapter_ConnectivityFailure(t *testing.T) {
	adapter := NewSelfAdapter("self")
	carrier := &LogisticsModel.Carrier{
		ID:     "carrier-1",
		Status: "active",
		Config: datatypes.JSON([]byte(`{"force_unreachable":true}`)),
	}
	err := adapter.TestConnectivity(context.Background(), carrier)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unreachable")
}

func TestSelfAdapter_CreateWaybill(t *testing.T) {
	adapter := NewSelfAdapter("sf")
	res, err := adapter.CreateWaybill(context.Background(), &LogisticsModel.Carrier{}, WaybillCreateInput{
		OrderID:     "order-1",
		ServiceCode: "std",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.WaybillNo)
	require.Equal(t, "created", res.Status)
	require.Equal(t, "sf", res.Metadata["provider"])
}
