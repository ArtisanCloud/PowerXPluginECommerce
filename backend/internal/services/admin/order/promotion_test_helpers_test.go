package order

import (
	"encoding/json"
	"testing"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedPromotion(t *testing.T, db *gorm.DB, row promotionmodel.Campaign) {
	t.Helper()
	require.NoError(t, db.Create(&row).Error)
}

func mustJSON(v any) []byte {
	raw, _ := json.Marshal(v)
	return raw
}
