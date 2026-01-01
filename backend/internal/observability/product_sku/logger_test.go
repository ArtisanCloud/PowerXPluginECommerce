package product_sku

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestEmitEventAddsCanonicalFields(t *testing.T) {
	base, hook := test.NewNullLogger()
	logger := NewLogger(logrus.NewEntry(base))

	logger.EmitEvent(Event{
		Action:        "bulk_task_submitted",
		TenantID:      "tenant-001",
		ActorID:       "user-123",
		TaskID:        "task-456",
		Duration:      150 * time.Millisecond,
		AffectedCount: 37,
		Metadata: map[string]any{
			"scope_size": 37,
			"action":     "should-ignore",
		},
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}

	entry := hook.LastEntry()
	if entry.Message != eventMessage {
		t.Fatalf("unexpected log message %q", entry.Message)
	}
	if got := entry.Data["action"]; got != "bulk_task_submitted" {
		t.Fatalf("expected action bulk_task_submitted, got %v", got)
	}
	if got := entry.Data["tenant_id"]; got != "tenant-001" {
		t.Fatalf("expected tenant_id tenant-001, got %v", got)
	}
	if got := entry.Data["actor_id"]; got != "user-123" {
		t.Fatalf("expected actor_id user-123, got %v", got)
	}
	if got := entry.Data["task_id"]; got != "task-456" {
		t.Fatalf("expected task_id task-456, got %v", got)
	}
	if got := entry.Data["scope_size"]; got != 37 {
		t.Fatalf("expected scope_size 37, got %v", got)
	}
	if got := entry.Data["duration_ms"]; got != int64(150) {
		t.Fatalf("expected duration_ms 150, got %v", got)
	}
	if got := entry.Data["affected_count"]; got != 37 {
		t.Fatalf("expected affected_count 37, got %v", got)
	}
	if _, ok := entry.Data["emitted_at"]; !ok {
		t.Fatalf("expected emitted_at field to be present")
	}
}

func TestEmitAuditIncludesAuditFields(t *testing.T) {
	base, hook := test.NewNullLogger()
	logger := NewLogger(logrus.NewEntry(base))

	logger.EmitAudit(AuditEvent{
		Action:     "bulk_task_approval_decided",
		TenantID:   "tenant-777",
		ActorID:    "approver-42",
		TargetType: "bulk_task",
		TargetID:   "task-777",
		Result:     "approved",
		Reason:     "threshold satisfied",
		Metadata: map[string]any{
			"approval_state": "approved",
		},
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}

	entry := hook.LastEntry()
	if entry.Message != auditEventMessage {
		t.Fatalf("unexpected audit log message %q", entry.Message)
	}
	if got := entry.Data["audit"]; got != true {
		t.Fatalf("expected audit flag true, got %v", got)
	}
	if got := entry.Data["actor_id"]; got != "approver-42" {
		t.Fatalf("expected actor_id approver-42, got %v", got)
	}
	if got := entry.Data["target_type"]; got != "bulk_task" {
		t.Fatalf("expected target_type bulk_task, got %v", got)
	}
	if got := entry.Data["approval_state"]; got != "approved" {
		t.Fatalf("expected approval_state approved, got %v", got)
	}
}
