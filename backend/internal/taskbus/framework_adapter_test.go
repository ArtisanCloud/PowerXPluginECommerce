package taskbus

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

func TestNormalizeHostBaseURL(t *testing.T) {
	got := normalizeHostBaseURL("http://localhost:8077/api/v1")
	if got != "http://localhost:8077" {
		t.Fatalf("unexpected base url: %s", got)
	}
}

func TestToFrameworkEventUsesContextTenant(t *testing.T) {
	client := &FrameworkClient{cfg: FrameworkClientConfig{SourcePlugin: "plugin.a", PayloadVersion: "v1"}}
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-from-ctx")
	ev, err := client.toFrameworkEvent(ctx, Event{Topic: "abc", Payload: map[string]any{"ok": true}, Metadata: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Meta.TenantUUID != "tenant-from-ctx" {
		t.Fatalf("unexpected tenant: %s", ev.Meta.TenantUUID)
	}
}
