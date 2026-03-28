package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCrossborderService_DocumentQuoteAndNormalize(t *testing.T) {
	db := setupBillingDB(t, "logistics_crossborder")
	ensureCrossborderTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-cross")
	svc := NewCrossborderService(&app.Deps{DB: db})

	doc, err := svc.UpsertDocument(ctx, "tenant-cross", UpsertCrossborderDocumentRequest{
		WaybillNo:   "WB-CROSS-1",
		DocType:     "invoice",
		DocNo:       "INV-001",
		CountryFrom: "CN",
		CountryTo:   "US",
		Status:      "validated",
	})
	require.NoError(t, err)
	require.Equal(t, "validated", doc.Status)

	rows, err := svc.ListDocuments(ctx, "tenant-cross", "WB-CROSS-1", 20)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	quote, idem, err := svc.QuoteTax(ctx, "tenant-cross", QuoteCrossborderTaxRequest{
		RequestKey:    "tax#1",
		WaybillNo:     "WB-CROSS-1",
		Destination:   "US",
		Currency:      "USD",
		DeclaredValue: 500,
		ShippingFee:   20,
		InsuranceFee:  5,
	})
	require.NoError(t, err)
	require.Equal(t, "created", idem)
	require.Equal(t, 0.0, quote.VatRate)
	require.True(t, quote.TotalTaxAmount >= 0)

	replayed, idem, err := svc.QuoteTax(ctx, "tenant-cross", QuoteCrossborderTaxRequest{
		RequestKey:    "tax#1",
		Destination:   "US",
		DeclaredValue: 800,
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", idem)
	require.Equal(t, quote.ID, replayed.ID)

	_, err = svc.UpsertTrackingMap(ctx, "tenant-cross", UpsertCrossborderTrackingMapRequest{
		Provider:         "dhl",
		ProviderStatus:   "customs_hold",
		NormalizedStatus: "exception",
		Priority:         10,
	})
	require.NoError(t, err)

	normalized, source, err := svc.NormalizeTrackingStatus(ctx, "tenant-cross", NormalizeCrossborderTrackingRequest{
		Provider:       "dhl",
		ProviderStatus: "customs_hold",
	})
	require.NoError(t, err)
	require.Equal(t, "exception", normalized)
	require.Equal(t, "mapping_table", source)

	fallback, fallbackSource, err := svc.NormalizeTrackingStatus(ctx, "tenant-cross", NormalizeCrossborderTrackingRequest{
		Provider:       "dhl",
		ProviderStatus: "in_transit",
	})
	require.NoError(t, err)
	require.Equal(t, "in_transit", fallback)
	require.Equal(t, "builtin", fallbackSource)
}

func TestCrossborderService_TaxBoundaryAndTenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_crossborder_tenant")
	ensureCrossborderTables(t, db)
	svc := NewCrossborderService(&app.Deps{DB: db})
	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-b")

	quote, _, err := svc.QuoteTax(ctxA, "tenant-a", QuoteCrossborderTaxRequest{
		RequestKey:    "tax#a",
		Destination:   "JP",
		DeclaredValue: 100,
		ShippingFee:   10,
		InsuranceFee:  5,
	})
	require.NoError(t, err)
	require.Equal(t, 0.0, quote.TotalTaxAmount)

	_, err = svc.UpsertDocument(ctxA, "tenant-a", UpsertCrossborderDocumentRequest{
		WaybillNo: "WB-A",
		DocType:   "invoice",
		DocNo:     "A-1",
	})
	require.NoError(t, err)

	rowsB, err := svc.ListDocuments(ctxB, "tenant-b", "", 20)
	require.NoError(t, err)
	require.Len(t, rowsB, 0)
}

func ensureCrossborderTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_crossborder_documents (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT NOT NULL,
		doc_type TEXT NOT NULL,
		doc_no TEXT NOT NULL,
		country_from TEXT NOT NULL DEFAULT '',
		country_to TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		validated_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_crossborder_doc ON logistics_crossborder_documents(tenant_uuid, waybill_no, doc_type)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_crossborder_tax_quotes (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT,
		destination_country TEXT NOT NULL,
		currency TEXT NOT NULL DEFAULT 'USD',
		declared_value NUMERIC NOT NULL DEFAULT 0,
		shipping_fee NUMERIC NOT NULL DEFAULT 0,
		insurance_fee NUMERIC NOT NULL DEFAULT 0,
		exemption_amount NUMERIC NOT NULL DEFAULT 0,
		duty_rate NUMERIC NOT NULL DEFAULT 0,
		vat_rate NUMERIC NOT NULL DEFAULT 0,
		duty_amount NUMERIC NOT NULL DEFAULT 0,
		vat_amount NUMERIC NOT NULL DEFAULT 0,
		total_tax_amount NUMERIC NOT NULL DEFAULT 0,
		quote_provider TEXT NOT NULL DEFAULT 'rule-engine',
		normalized_status TEXT NOT NULL DEFAULT 'estimated',
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_crossborder_tax_quote ON logistics_crossborder_tax_quotes(tenant_uuid, request_key)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_crossborder_tracking_maps (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		provider TEXT NOT NULL,
		provider_status TEXT NOT NULL,
		normalized_status TEXT NOT NULL,
		description TEXT,
		priority INTEGER NOT NULL DEFAULT 100,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_crossborder_tracking_map ON logistics_crossborder_tracking_maps(tenant_uuid, provider, provider_status)`).Error)
}
