package payments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestManualReviewService_ApproveUpdatesOrder(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createManualReviewTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const reviewerID = "2002"

	require.NoError(t, seedOrder(t, db, tenant, "ord-1", "O20260101010101ABCDEF", "pending_payment", "CNY", 10000))
	reviewID := seedManualReview(t, db, tenant, "ord-1", "O20260101010101ABCDEF", adminID)

	svc := NewManualReviewService(&app.Deps{DB: db})
	resp, err := svc.ApproveReview(ctx, tenant, reviewerID, reviewID, ManualPaymentReviewRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "approved", resp.Status)
	require.NotNil(t, resp.TransactionID)

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM orders WHERE tenant_uuid = ? AND id = ?`, tenant, "ord-1").Scan(&status).Error)
	require.Equal(t, "paid", status)

	var txCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM payment_transactions WHERE tenant_uuid = ? AND order_id = ?`, tenant, "ord-1").Scan(&txCount).Error)
	require.Equal(t, int64(1), txCount)

	_, err = svc.ApproveReview(ctx, tenant, reviewerID, reviewID, ManualPaymentReviewRequest{})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrManualReviewInvalidStatus))
}

func TestManualReviewService_RejectKeepsOrderStatus(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createManualReviewTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const reviewerID = "2002"

	require.NoError(t, seedOrder(t, db, tenant, "ord-2", "O20260101010101ABCXYZ", "pending_payment", "CNY", 5000))
	reviewID := seedManualReview(t, db, tenant, "ord-2", "O20260101010101ABCXYZ", adminID)

	svc := NewManualReviewService(&app.Deps{DB: db})
	resp, err := svc.RejectReview(ctx, tenant, reviewerID, reviewID, ManualPaymentReviewRequest{Reason: "金额不符"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "rejected", resp.Status)

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM orders WHERE tenant_uuid = ? AND id = ?`, tenant, "ord-2").Scan(&status).Error)
	require.Equal(t, "pending_payment", status)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	return db
}

func createManualReviewTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_no TEXT NOT NULL,
			status TEXT NOT NULL,
			currency TEXT NOT NULL,
			total_amount INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_events (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			operator_type TEXT NOT NULL,
			operator TEXT,
			payload TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			transaction_no TEXT NOT NULL,
			order_id TEXT,
			order_no TEXT,
			provider_id INTEGER NOT NULL DEFAULT 0,
			pay_method TEXT NOT NULL,
			amount_total INTEGER NOT NULL,
			amount_currency TEXT NOT NULL,
			fee_amount INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			completed_at DATETIME,
			failure_reason TEXT,
			risk_flag BOOLEAN NOT NULL DEFAULT 0,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_manual_reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			order_no TEXT NOT NULL,
			transaction_id INTEGER,
			provider_id INTEGER NOT NULL DEFAULT 0,
			pay_method TEXT NOT NULL,
			amount_minor INTEGER NOT NULL DEFAULT 0,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			submitted_by TEXT NOT NULL,
			submitted_at DATETIME,
			reviewed_by TEXT,
			reviewed_at DATETIME,
			review_reason TEXT,
			proof_no TEXT,
			note TEXT,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_manual_review_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			review_id INTEGER NOT NULL,
			order_id TEXT NOT NULL,
			order_no TEXT NOT NULL,
			pay_method TEXT NOT NULL,
			amount_minor INTEGER NOT NULL DEFAULT 0,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			action TEXT NOT NULL,
			submitted_by TEXT NOT NULL,
			submitted_at DATETIME,
			reviewed_by TEXT,
			reviewed_at DATETIME,
			review_reason TEXT,
			proof_no TEXT,
			note TEXT,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}

func seedOrder(t *testing.T, db *gorm.DB, tenant, orderID, orderNo, status, currency string, total int64) error {
	t.Helper()
	return db.Exec(`INSERT INTO orders (id, tenant_uuid, order_no, status, currency, total_amount, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, orderID, tenant, orderNo, status, currency, total, time.Now().UTC(), time.Now().UTC()).Error
}

func seedManualReview(t *testing.T, db *gorm.DB, tenant, orderID, orderNo, submitter string) uint64 {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO payment_manual_reviews (tenant_uuid, order_id, order_no, provider_id, pay_method, amount_minor, currency, status, submitted_by, submitted_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenant, orderID, orderNo, int64(0), "bank_transfer", int64(10000), "CNY", "pending_review", submitter, now, now, now).Error)
	var id uint64
	require.NoError(t, db.Raw(`SELECT id FROM payment_manual_reviews WHERE tenant_uuid = ? AND order_id = ? ORDER BY id DESC LIMIT 1`, tenant, orderID).Scan(&id).Error)
	return id
}
