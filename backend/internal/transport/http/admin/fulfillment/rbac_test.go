package fulfillment

import "testing"

func TestRBACEntries_FulfillmentRoutes(t *testing.T) {
	entries := RBACEntries("/api/v1")

	if len(entries) != 17 {
		t.Fatalf("expected 17 fulfillment RBAC entries, got %d", len(entries))
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

	waveKey := "PATCH:/api/v1/admin/fulfillment/waves/:id/advance"
	wavePerm, ok := entries[waveKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", waveKey)
	}
	if wavePerm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", wavePerm.Action)
	}
	if wavePerm.Resource != "com.powerx.plugins.ecommerce:fulfillment.wave" {
		t.Fatalf("unexpected resource %s", wavePerm.Resource)
	}

	strategyKey := "POST:/api/v1/admin/fulfillment/wave-strategies/preview"
	strategyPerm, ok := entries[strategyKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", strategyKey)
	}
	if strategyPerm.Resource != "com.powerx.plugins.ecommerce:fulfillment.wave.strategy" {
		t.Fatalf("unexpected resource %s", strategyPerm.Resource)
	}

	warehouseKey := "POST:/api/v1/admin/fulfillment/warehouse/outbounds/:id/execute"
	warehousePerm, ok := entries[warehouseKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", warehouseKey)
	}
	if warehousePerm.Resource != "com.powerx.plugins.ecommerce:fulfillment.warehouse" {
		t.Fatalf("unexpected resource %s", warehousePerm.Resource)
	}
}
