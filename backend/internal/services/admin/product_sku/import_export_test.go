package product_sku

import (
	"strings"
	"testing"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/stretchr/testify/require"
)

func TestParseSkuImportRows(t *testing.T) {
	content := "sku_code,price,inventory\nSKU-001,12.5,5\nSKU-002,,10\n"
	rows, err := parseSkuImportRows([]byte(content))
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "SKU-001", rows[0].SKUCode)
	require.NotNil(t, rows[0].Price)
	require.Equal(t, 12.5, *rows[0].Price)
	require.NotNil(t, rows[1].Inventory)
	require.Equal(t, int64(10), *rows[1].Inventory)
}

func TestParseSkuImportRowsMissingHeader(t *testing.T) {
	_, err := parseSkuImportRows([]byte("sku,status\n1,ready\n"))
	require.Error(t, err)
}

func TestBuildSkuExportCSV(t *testing.T) {
	csv := buildSkuExportCSV([]productskumodel.ProductSKU{
		{SKUCode: "S-1", Status: "draft", Barcode: "B1", MinOrderQty: 5},
	})
	require.True(t, strings.Contains(csv, "sku_code"))
	require.True(t, strings.Contains(csv, "S-1"))
}
