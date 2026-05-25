package logger

import (
	"os"
	"path/filepath"
	"testing"

	runtimelogging "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/logging"
)

func TestInitWithHostModeKeepsFileOutputForLocalMode(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "plugin.log")

	InitWithHostMode("info", "json", "file", logPath, 100, 3, 28, true, false)
	t.Cleanup(func() {
		runtimelogging.SetHostModeOverride(false)
	})

	WithField("component", "logger_test").Info("local file log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected local file log to be created: %v", err)
	}
}

func TestInitWithHostModeForcesStdoutJsonForHostMode(t *testing.T) {
	policy := runtimelogging.ResolveWithHostDefaults(runtimelogging.Policy{
		Mode:   runtimelogging.ModeHost,
		Format: "text",
		Sinks:  []runtimelogging.SinkType{runtimelogging.SinkFile},
	})

	if policy.Format != "json" {
		t.Fatalf("host logging format = %q, want json", policy.Format)
	}
	if runtimelogging.PrimaryOutput(policy) != string(runtimelogging.SinkStdout) {
		t.Fatalf("host logging primary output = %q, want stdout", runtimelogging.PrimaryOutput(policy))
	}
}
