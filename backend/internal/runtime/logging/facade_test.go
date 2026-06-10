package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestFacadeEmitMergesContextAndCallFields(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	ctx := ContextWithFields(context.Background(), Fields{
		FieldPluginID:   "plugin.test",
		FieldTenantUUID: "tenant-1",
	})
	FromContext(ctx).With(Fields{
		FieldComponent: "runtime.test",
	}).Emit("info", "probe ok", Fields{
		FieldTraceID: "trace-1",
	})

	var row map[string]any
	if err := json.Unmarshal(buf.Bytes(), &row); err != nil {
		t.Fatalf("decode log row: %v", err)
	}
	if row[FieldPluginID] != "plugin.test" {
		t.Fatalf("plugin_id missing: %#v", row)
	}
	if row[FieldTenantUUID] != "tenant-1" {
		t.Fatalf("tenant_uuid missing: %#v", row)
	}
	if row[FieldComponent] != "runtime.test" {
		t.Fatalf("component missing: %#v", row)
	}
	if row[FieldTraceID] != "trace-1" {
		t.Fatalf("trace_id missing: %#v", row)
	}
	if row[FieldStatus] != "succeeded" {
		t.Fatalf("default status missing: %#v", row)
	}
}
