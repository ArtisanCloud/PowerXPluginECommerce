package fulfillment

import "testing"

func TestRBACEntries_FulfillmentRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) != 5 {
		t.Fatalf("expected 5 fulfillment RBAC entries, got %d", len(entries))
	}

	key := "PATCH:/api/v1/admin/fulfillment/tasks/:id/complete"
	perm, ok := entries[key]
	if !ok {
		t.Fatalf("missing rbac entry: %s", key)
	}
	if perm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", perm.Action)
	}
	if perm.Resource != "com.powerx.plugins.ecommerce:fulfillment.task" {
		t.Fatalf("unexpected resource %s", perm.Resource)
	}
}
