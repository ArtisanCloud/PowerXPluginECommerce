package logistics

import "testing"

func TestRBACEntries_LogisticsRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) < 13 {
		t.Fatalf("expected at least 13 logistics RBAC entries, got %d", len(entries))
	}

	key := "POST:/api/v1/admin/logistics/waybills"
	perm, ok := entries[key]
	if !ok {
		t.Fatalf("missing rbac entry: %s", key)
	}
	if perm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", perm.Action)
	}
	if perm.Resource != "com.powerx.plugins.ecommerce:logistics.waybill" {
		t.Fatalf("unexpected resource %s", perm.Resource)
	}
}
