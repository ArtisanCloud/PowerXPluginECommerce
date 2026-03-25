package reverse

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestEmitterEmit_ContainsStructuredFields(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.Emit(Event{
		Action:      "reverse.track.append",
		TenantUUID:  "tenant-c",
		RequestID:   "req-c",
		WaybillNo:   "rwb-c",
		AfterSaleID: "as-c",
		Status:      "received",
		Result:      "ok",
	})

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}
	entry := hook.LastEntry()
	if entry.Message != eventMsg {
		t.Fatalf("unexpected message %s", entry.Message)
	}
	if entry.Data["tenant_uuid"] != "tenant-c" {
		t.Fatalf("expected tenant_uuid tenant-c, got %v", entry.Data["tenant_uuid"])
	}
	if entry.Data["request_id"] != "req-c" {
		t.Fatalf("expected request_id req-c, got %v", entry.Data["request_id"])
	}
	if entry.Data["waybill_no"] != "rwb-c" {
		t.Fatalf("expected waybill_no rwb-c, got %v", entry.Data["waybill_no"])
	}
}

func TestEmitterEmitAudit_ContainsAuditFlag(t *testing.T) {
	base, hook := test.NewNullLogger()
	emitter := NewEmitter(logrus.NewEntry(base))

	emitter.EmitAudit(AuditEvent{
		Action:     "reverse.warehouse.result",
		TenantUUID: "tenant-c",
		ActorID:    "user-c",
		TargetID:   "rwb-c",
		Result:     "accepted",
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
	if entry.Data["tenant_uuid"] != "tenant-c" {
		t.Fatalf("expected tenant_uuid tenant-c, got %v", entry.Data["tenant_uuid"])
	}
}
