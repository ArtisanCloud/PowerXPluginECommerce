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

	routingKey := "POST:/api/v1/admin/logistics/routing/preview"
	routingPerm, ok := entries[routingKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", routingKey)
	}
	if routingPerm.Resource != "com.powerx.plugins.ecommerce:logistics.routing" {
		t.Fatalf("unexpected resource %s", routingPerm.Resource)
	}

	redeliveryKey := "POST:/api/v1/admin/logistics/redelivery/tasks/:id/redispatch"
	redeliveryPerm, ok := entries[redeliveryKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", redeliveryKey)
	}
	if redeliveryPerm.Resource != "com.powerx.plugins.ecommerce:logistics.redelivery" {
		t.Fatalf("unexpected resource %s", redeliveryPerm.Resource)
	}

	syncTrackKey := "POST:/api/v1/admin/logistics/waybills/:id/sync-track"
	syncTrackPerm, ok := entries[syncTrackKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", syncTrackKey)
	}
	if syncTrackPerm.Resource != "com.powerx.plugins.ecommerce:logistics.waybill" {
		t.Fatalf("unexpected resource %s", syncTrackPerm.Resource)
	}

	riskKey := "POST:/api/v1/admin/logistics/risk/hits/:id/release"
	riskPerm, ok := entries[riskKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", riskKey)
	}
	if riskPerm.Resource != "com.powerx.plugins.ecommerce:logistics.risk" {
		t.Fatalf("unexpected resource %s", riskPerm.Resource)
	}

	trackingSyncKey := "POST:/api/v1/admin/logistics/tracking-sync/jobs"
	trackingSyncPerm, ok := entries[trackingSyncKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", trackingSyncKey)
	}
	if trackingSyncPerm.Resource != "com.powerx.plugins.ecommerce:logistics.tracking_sync" {
		t.Fatalf("unexpected resource %s", trackingSyncPerm.Resource)
	}

	gatewayKey := "GET:/api/v1/admin/logistics/gateway/health"
	gatewayPerm, ok := entries[gatewayKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", gatewayKey)
	}
	if gatewayPerm.Action != "read" {
		t.Fatalf("expected read action, got %s", gatewayPerm.Action)
	}
	costKey := "GET:/api/v1/admin/logistics/gateway/costs"
	costPerm, ok := entries[costKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", costKey)
	}
	if costPerm.Action != "read" {
		t.Fatalf("expected read action, got %s", costPerm.Action)
	}
	alertKey := "GET:/api/v1/admin/logistics/gateway/cost-alerts"
	alertPerm, ok := entries[alertKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", alertKey)
	}
	if alertPerm.Resource != "com.powerx.plugins.ecommerce:logistics.gateway" {
		t.Fatalf("unexpected resource %s", alertPerm.Resource)
	}

	scheduleKey := "POST:/api/v1/admin/logistics/tracking-sync/schedules"
	schedulePerm, ok := entries[scheduleKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", scheduleKey)
	}
	if schedulePerm.Resource != "com.powerx.plugins.ecommerce:logistics.tracking_sync" {
		t.Fatalf("unexpected resource %s", schedulePerm.Resource)
	}

	failureCompensateKey := "POST:/api/v1/admin/logistics/gateway/failures/:id/compensate"
	failureCompensatePerm, ok := entries[failureCompensateKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", failureCompensateKey)
	}
	if failureCompensatePerm.Action != "manage" {
		t.Fatalf("expected manage action, got %s", failureCompensatePerm.Action)
	}

	orchestrationKey := "POST:/api/v1/admin/logistics/exceptions/orchestration/execute"
	orchestrationPerm, ok := entries[orchestrationKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", orchestrationKey)
	}
	if orchestrationPerm.Resource != "com.powerx.plugins.ecommerce:logistics.orchestration" {
		t.Fatalf("unexpected resource %s", orchestrationPerm.Resource)
	}

	addressCheckKey := "POST:/api/v1/admin/logistics/address-validation/check"
	addressCheckPerm, ok := entries[addressCheckKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", addressCheckKey)
	}
	if addressCheckPerm.Resource != "com.powerx.plugins.ecommerce:logistics.address_validation" {
		t.Fatalf("unexpected resource %s", addressCheckPerm.Resource)
	}

	optimizerKey := "POST:/api/v1/admin/logistics/routing/optimizer/simulate"
	optimizerPerm, ok := entries[optimizerKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", optimizerKey)
	}
	if optimizerPerm.Resource != "com.powerx.plugins.ecommerce:logistics.routing_optimizer" {
		t.Fatalf("unexpected resource %s", optimizerPerm.Resource)
	}

	settlementKey := "POST:/api/v1/admin/logistics/settlement/batches/:id/confirm"
	settlementPerm, ok := entries[settlementKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", settlementKey)
	}
	if settlementPerm.Resource != "com.powerx.plugins.ecommerce:logistics.settlement" {
		t.Fatalf("unexpected resource %s", settlementPerm.Resource)
	}

	controlTowerKey := "GET:/api/v1/admin/logistics/control-tower/overview"
	controlTowerPerm, ok := entries[controlTowerKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", controlTowerKey)
	}
	if controlTowerPerm.Resource != "com.powerx.plugins.ecommerce:logistics.control_tower" {
		t.Fatalf("unexpected resource %s", controlTowerPerm.Resource)
	}

	allocationKey := "POST:/api/v1/admin/logistics/allocation/allocate"
	allocationPerm, ok := entries[allocationKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", allocationKey)
	}
	if allocationPerm.Resource != "com.powerx.plugins.ecommerce:logistics.allocation" {
		t.Fatalf("unexpected resource %s", allocationPerm.Resource)
	}

	lastmileKey := "POST:/api/v1/admin/logistics/lastmile-recovery/execute"
	lastmilePerm, ok := entries[lastmileKey]
	if !ok {
		t.Fatalf("missing rbac entry: %s", lastmileKey)
	}
	if lastmilePerm.Resource != "com.powerx.plugins.ecommerce:logistics.lastmile_recovery" {
		t.Fatalf("unexpected resource %s", lastmilePerm.Resource)
	}
}
