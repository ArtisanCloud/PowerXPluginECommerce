package channel_master

import (
	"context"
	"errors"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TaskLinkInput captures linking payload.
type TaskLinkInput struct {
	TaskID     string `json:"task_id"`
	TaskSource string `json:"task_source"`
	Status     string `json:"status"`
	Note       string `json:"note"`
}

// NoteInput stores note body/visibility.
type NoteInput struct {
	Body       string `json:"body"`
	Visibility string `json:"visibility"`
}

// TaskNoteService manages remediation task links and notes.
type TaskNoteService struct {
	taskRepo *channelrepo.ChannelTaskLinkRepository
	noteRepo *channelrepo.ChannelNoteRepository
	logger   *logrus.Entry
	audit    AuditEmitter
}

// NewTaskNoteService wires dependencies.
func NewTaskNoteService(deps *app.Deps, audit AuditEmitter) *TaskNoteService {
	if deps == nil || deps.DB == nil {
		panic("task note service requires database dependency")
	}
	logger := deps.RuntimeLogger(nil, "channel-task-note-service", nil)
	if audit == nil {
		audit = channelobs.NewAuditEmitter(logger)
	}
	return &TaskNoteService{
		taskRepo: channelrepo.NewChannelTaskLinkRepository(deps.DB),
		noteRepo: channelrepo.NewChannelNoteRepository(deps.DB),
		logger:   logger,
		audit:    audit,
	}
}

// LinkTask creates association to remediation task.
func (s *TaskNoteService) LinkTask(ctx context.Context, channelID string, input TaskLinkInput) (*channelmodel.ChannelTaskLink, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	link := &channelmodel.ChannelTaskLink{
		ID:         uuid.NewString(),
		ChannelID:  strings.TrimSpace(channelID),
		TaskID:     input.TaskID,
		TaskSource: strings.ToLower(strings.TrimSpace(input.TaskSource)),
		Status:     input.Status,
		Note:       input.Note,
		LinkedBy:   actorFromContext(ctx),
		LinkedAt:   time.Now().UTC(),
	}
	saved, err := s.taskRepo.Create(ctx, link)
	if err != nil {
		return nil, err
	}
	_ = s.audit.EmitChannelAudit(ctx, "channel.task.linked", map[string]any{
		"channel_id": channelID,
		"task_id":    input.TaskID,
	})
	return saved, nil
}

// UpdateTask mutates link status/note.
func (s *TaskNoteService) UpdateTask(ctx context.Context, linkID string, status string, note string) error {
	if strings.TrimSpace(linkID) == "" {
		return errors.New("task link id required")
	}
	if strings.TrimSpace(status) == "" {
		return errors.New("status required")
	}
	if err := s.taskRepo.UpdateStatus(ctx, linkID, strings.ToLower(strings.TrimSpace(status)), strings.TrimSpace(note)); err != nil {
		return err
	}
	_ = s.audit.EmitChannelAudit(ctx, "channel.task.updated", map[string]any{
		"task_link_id": linkID,
		"status":       status,
	})
	return nil
}

// RemoveTask deletes association.
func (s *TaskNoteService) RemoveTask(ctx context.Context, linkID string) error {
	if strings.TrimSpace(linkID) == "" {
		return errors.New("task link id required")
	}
	if err := s.taskRepo.Delete(ctx, linkID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	_ = s.audit.EmitChannelAudit(ctx, "channel.task.removed", map[string]any{"task_link_id": linkID})
	return nil
}

// ListTasks returns associated tasks.
func (s *TaskNoteService) ListTasks(ctx context.Context, channelID string) ([]*channelmodel.ChannelTaskLink, error) {
	return s.taskRepo.ListByChannel(ctx, channelID)
}

// AddNote stores operator remarks.
func (s *TaskNoteService) AddNote(ctx context.Context, channelID string, input NoteInput) (*channelmodel.ChannelNote, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, errors.New("channel id required")
	}
	if err := input.validate(); err != nil {
		return nil, err
	}
	note := &channelmodel.ChannelNote{
		ID:         uuid.NewString(),
		ChannelID:  channelID,
		AuthorUUID: actorFromContext(ctx),
		Visibility: strings.ToLower(strings.TrimSpace(input.Visibility)),
		Body:       input.Body,
	}
	saved, err := s.noteRepo.Create(ctx, note)
	if err != nil {
		return nil, err
	}
	_ = s.audit.EmitChannelAudit(ctx, "channel.note.created", map[string]any{"channel_id": channelID})
	return saved, nil
}

// ListNotes returns recent notes.
func (s *TaskNoteService) ListNotes(ctx context.Context, channelID string, limit int) ([]*channelmodel.ChannelNote, error) {
	return s.noteRepo.ListByChannel(ctx, channelID, limit)
}

func (in *TaskLinkInput) validate() error {
	if strings.TrimSpace(in.TaskID) == "" {
		return errors.New("task_id required")
	}
	if strings.TrimSpace(in.TaskSource) == "" {
		in.TaskSource = "task_center"
	}
	if strings.TrimSpace(in.Status) == "" {
		in.Status = "open"
	}
	in.TaskID = strings.TrimSpace(in.TaskID)
	in.TaskSource = strings.ToLower(strings.TrimSpace(in.TaskSource))
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	return nil
}

func (in *NoteInput) validate() error {
	if strings.TrimSpace(in.Body) == "" {
		return errors.New("note body required")
	}
	if strings.TrimSpace(in.Visibility) == "" {
		in.Visibility = "team"
	}
	in.Body = strings.TrimSpace(in.Body)
	in.Visibility = strings.ToLower(strings.TrimSpace(in.Visibility))
	return nil
}
