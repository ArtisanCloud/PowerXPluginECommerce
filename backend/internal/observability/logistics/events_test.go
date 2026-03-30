package logistics

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestEmitterEmit_ContainsStructuredFields(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.Emit(Event{
		Action:     "waybill.track.append",
		TenantUUID: "tenant-a",
		RequestID:  "req-a",
		WaybillNo:  "wb-a",
		Status:     "in_transit",
		Result:     "ok",
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}
	entry := hook.LastEntry()
	if entry.Message != eventMsg {
		t.Fatalf("unexpected message %s", entry.Message)
	}
	if entry.Data["tenant_uuid"] != "tenant-a" {
		t.Fatalf("expected tenant_uuid tenant-a, got %v", entry.Data["tenant_uuid"])
	}
	if entry.Data["request_id"] != "req-a" {
		t.Fatalf("expected request_id req-a, got %v", entry.Data["request_id"])
	}
	if entry.Data["waybill_no"] != "wb-a" {
		t.Fatalf("expected waybill_no wb-a, got %v", entry.Data["waybill_no"])
	}
}

func TestEmitterEmitAudit_ContainsAuditFlag(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.EmitAudit(AuditEvent{
		Action:     "carrier.test",
		TenantUUID: "tenant-a",
		ActorID:    "user-a",
		TargetID:   "carrier-a",
		Result:     "reachable=true",
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
	if entry.Data["tenant_uuid"] != "tenant-a" {
		t.Fatalf("expected tenant_uuid tenant-a, got %v", entry.Data["tenant_uuid"])
	}
}
