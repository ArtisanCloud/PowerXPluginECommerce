package logistics

import "testing"

func TestRBACEntries_LogisticsRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) < 18 {
		t.Fatalf("expected at least 18 logistics RBAC entries, got %d", len(entries))
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

	exportKey := "GET:/api/v1/admin/logistics/billing/export"
	exportPerm, ok := entries[exportKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", exportKey)
	}
	if exportPerm.Action != "export" {
		t.Fatalf("expected export action, got %s", exportPerm.Action)
	}

	printKey := "POST:/api/v1/admin/logistics/labels/prints"
	printPerm, ok := entries[printKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", printKey)
	}
	if printPerm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", printPerm.Action)
	}
	if printPerm.Resource != "com.powerx.plugins.ecommerce:logistics.label_print" {
		t.Fatalf("unexpected resource %s", printPerm.Resource)
	}

	slaKey := "GET:/api/v1/admin/logistics/sla/dashboard"
	slaPerm, ok := entries[slaKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", slaKey)
	}
	if slaPerm.Action != "read" {
		t.Fatalf("expected read action, got %s", slaPerm.Action)
	}
	if slaPerm.Resource != "com.powerx.plugins.ecommerce:logistics.sla" {
		t.Fatalf("unexpected resource %s", slaPerm.Resource)
	}

	quoteKey := "POST:/api/v1/admin/logistics/templates/:id/quote"
	quotePerm, ok := entries[quoteKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", quoteKey)
	}
	if quotePerm.Action != "read" {
		t.Fatalf("expected read action, got %s", quotePerm.Action)
	}

	caseKey := "PATCH:/api/v1/admin/logistics/billing/cases/:id/transition"
	casePerm, ok := entries[caseKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", caseKey)
	}
	if casePerm.Resource != "com.powerx.plugins.ecommerce:logistics.billing.case" {
		t.Fatalf("unexpected resource %s", casePerm.Resource)
	}

	notifyKey := "POST:/api/v1/admin/logistics/notifications/send"
	notifyPerm, ok := entries[notifyKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", notifyKey)
	}
	if notifyPerm.Resource != "com.powerx.plugins.ecommerce:logistics.notification" {
		t.Fatalf("unexpected resource %s", notifyPerm.Resource)
	}
}
