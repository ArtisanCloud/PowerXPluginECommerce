package channel_master

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTaskNoteService_LinkAndList(t *testing.T) {
	svc, ctx, db := newTestTaskNoteService(t)
	_, err := svc.LinkTask(ctx, "channel-1", TaskLinkInput{
		TaskID:     "TASK-100",
		TaskSource: "task_center",
		Status:     "open",
		Note:       "需要补充素材",
	})
	require.NoError(t, err)

	tasks, err := svc.ListTasks(ctx, "channel-1")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, "TASK-100", tasks[0].TaskID)

	require.NoError(t, svc.UpdateTask(ctx, tasks[0].ID, "done", "已完成"))

	var stored struct {
		Status     string
		ResolvedAt *time.Time
	}
	require.NoError(t, db.WithContext(ctx).
		Table("channel_task_links").
		Select("status", "resolved_at").
		Where("id = ?", tasks[0].ID).
		Scan(&stored).Error)
	require.Equal(t, "done", stored.Status)
	require.NotNil(t, stored.ResolvedAt)

	require.NoError(t, svc.RemoveTask(ctx, tasks[0].ID))
}

func TestTaskNoteService_AddNote(t *testing.T) {
	svc, ctx, _ := newTestTaskNoteService(t)
	note, err := svc.AddNote(ctx, "channel-2", NoteInput{
		Body:       "等待品牌确认物料",
		Visibility: "team",
	})
	require.NoError(t, err)
	require.Equal(t, "channel-2", note.ChannelID)

	notes, err := svc.ListNotes(ctx, "channel-2", 10)
	require.NoError(t, err)
	require.Len(t, notes, 1)
	require.Equal(t, "等待品牌确认物料", notes[0].Body)
}

func newTestTaskNoteService(t *testing.T) (*TaskNoteService, context.Context, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS channel_task_links (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT,
			channel_id TEXT,
			task_id TEXT,
			task_source TEXT,
			status TEXT,
			note TEXT,
			linked_by TEXT,
			linked_at DATETIME,
			resolved_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS channel_notes (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT,
			channel_id TEXT,
			author_uuid TEXT,
			visibility TEXT,
			body TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
	deps := &app.Deps{
		DB: db,
		Config: &config.Config{
			Server: &config.ServerConfig{SecretKey: "task-note-test"},
		},
	}
	audit := channelobs.NewAuditEmitter(nil)
	svc := NewTaskNoteService(deps, audit)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-task-note")
	return svc, ctx, db
}
