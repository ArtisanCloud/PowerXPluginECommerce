package reverse

import "testing"

func TestRBACEntries_ReverseRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) != 8 {
		t.Fatalf("expected 8 reverse RBAC entries, got %d", len(entries))
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

	inspectionKey := "POST:/api/v1/admin/reverse/waybills/:id/inspection"
	inspectionPerm, ok := entries[inspectionKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", inspectionKey)
	}
	if inspectionPerm.Resource != "com.powerx.plugins.ecommerce:reverse.inspection" {
		t.Fatalf("unexpected inspection resource %s", inspectionPerm.Resource)
	}
}
