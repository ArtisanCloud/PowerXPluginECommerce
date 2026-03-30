package integrations

import (
	"testing"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestConfigService_ResolveProviderAndServiceCode(t *testing.T) {
	svc := NewConfigService()
	carrier := &LogisticsModel.Carrier{
		Type: "self",
		Config: datatypes.JSON([]byte(`{
			"provider":"sf",
			"service_code_map":{"std":"SF_STD"}
		}`)),
	}
	require.Equal(t, "sf", svc.ResolveProvider(carrier))
	require.Equal(t, "SF_STD", svc.ResolveServiceCode(carrier, "std"))
	require.Equal(t, "next_day", svc.ResolveServiceCode(carrier, "next_day"))

	carrier.Config = datatypes.JSON([]byte(`{"provider":"jd"}`))
	require.Equal(t, "jd", svc.ResolveProvider(carrier))
}
