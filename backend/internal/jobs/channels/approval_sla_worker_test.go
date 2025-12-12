package channels

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type mockReminder struct {
	called []string
	err    error
}

func (m *mockReminder) Send(ctx context.Context, record productmodel.SPUApprovalRecord) error {
	m.called = append(m.called, record.ID)
	return m.err
}

func TestApprovalSLAWorkerEscalatesAndNotifies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	models.ForceSchemaForTests("")
	require.NoError(t, db.Exec(`CREATE TABLE product_spu_approvals (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		spu_id TEXT,
		version_id TEXT,
		chain_order INTEGER,
		role TEXT,
		status TEXT,
		comment TEXT,
		sla_due_at DATETIME,
		acted_by TEXT,
		acted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error)

	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	require.NoError(t, db.Create(&productmodel.SPUApprovalRecord{
		ID:         "apr-overdue",
		TenantUUID: "tenant",
		SPUID:      "spu-1",
		VersionID:  "ver-1",
		ChainOrder: 1,
		Role:       "qc",
		Status:     "pending",
		SLADueAt:   &past,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)

	require.NoError(t, db.Create(&productmodel.SPUApprovalRecord{
		ID:         "apr-future",
		TenantUUID: "tenant",
		SPUID:      "spu-1",
		VersionID:  "ver-1",
		ChainOrder: 2,
		Role:       "legal",
		Status:     "pending",
		SLADueAt:   &future,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)

	notifier := &mockReminder{}
	worker := NewApprovalSLAWorker(db, notifier, nil).WithBatchSize(10).WithClock(func() time.Time { return now })

	count, err := worker.Process(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, []string{"apr-overdue"}, notifier.called)

	var overdue productmodel.SPUApprovalRecord
	require.NoError(t, db.Where("id = ?", "apr-overdue").First(&overdue).Error)
	require.Equal(t, "escalated", overdue.Status)

	var pending productmodel.SPUApprovalRecord
	require.NoError(t, db.Where("id = ?", "apr-future").First(&pending).Error)
	require.Equal(t, "pending", pending.Status)
}
