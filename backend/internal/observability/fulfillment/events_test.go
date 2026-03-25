package fulfillment

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestEmitterEmit_ContainsStructuredFields(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.Emit(Event{
		Action:      "task.complete",
		TenantUUID:  "tenant-b",
		RequestID:   "req-b",
		TaskID:      "task-b",
		ExceptionID: "ex-b",
		Status:      "completed",
		Result:      "ok",
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}
	entry := hook.LastEntry()
	if entry.Message != eventMsg {
		t.Fatalf("unexpected message %s", entry.Message)
	}
	if entry.Data["tenant_uuid"] != "tenant-b" {
		t.Fatalf("expected tenant_uuid tenant-b, got %v", entry.Data["tenant_uuid"])
	}
	if entry.Data["request_id"] != "req-b" {
		t.Fatalf("expected request_id req-b, got %v", entry.Data["request_id"])
	}
	if entry.Data["task_id"] != "task-b" {
		t.Fatalf("expected task_id task-b, got %v", entry.Data["task_id"])
	}
}

func TestEmitterEmitAudit_ContainsAuditFlag(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.EmitAudit(AuditEvent{
		Action:     "exception.create",
		TenantUUID: "tenant-b",
		ActorID:    "user-b",
		TargetID:   "exception-b",
		Result:     "created",
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}
	entry := hook.LastEntry()
	if entry.Message != auditMsg {
		t.Fatalf("unexpected message %s", entry.Message)
	}
	if entry.Data["audit"] != true {
		t.Fatalf("expected audit=true, got %v", entry.Data["audit"])
	}
	if entry.Data["tenant_uuid"] != "tenant-b" {
		t.Fatalf("expected tenant_uuid tenant-b, got %v", entry.Data["tenant_uuid"])
	}
}
