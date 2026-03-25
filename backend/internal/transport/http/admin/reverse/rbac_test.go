package reverse

import "testing"

func TestRBACEntries_ReverseRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) != 5 {
		t.Fatalf("expected 5 reverse RBAC entries, got %d", len(entries))
	}

	key := "POST:/api/v1/admin/reverse/waybills/:id/warehouse-result"
	perm, ok := entries[key]
	if !ok {
		t.Fatalf("missing rbac entry: %s", key)
	}
	if perm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", perm.Action)
	}
	if perm.Resource != "com.powerx.plugins.ecommerce:reverse.waybill" {
		t.Fatalf("unexpected resource %s", perm.Resource)
	}
}
