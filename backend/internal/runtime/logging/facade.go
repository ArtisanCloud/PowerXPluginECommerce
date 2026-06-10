package logging

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

type contextKey struct{}

// Logger is the runtime logging facade used by new runtime code.
type Logger struct {
	ctx    context.Context
	fields Fields
}

// FromContext returns a runtime logger seeded with fields stored on ctx.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	fields := Fields{}
	if existing, ok := ctx.Value(contextKey{}).(Fields); ok {
		for k, v := range existing {
			fields[k] = v
		}
	}
	return Logger{ctx: ctx, fields: fields}
}

// ContextWithFields stores runtime log fields on a context for downstream logs.
func ContextWithFields(ctx context.Context, fields Fields) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	merged := Fields{}
	if existing, ok := ctx.Value(contextKey{}).(Fields); ok {
		for k, v := range existing {
			merged[k] = v
		}
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, contextKey{}, merged)
}

// With returns a copy of the logger enriched with fields.
func (l Logger) With(fields Fields) Logger {
	next := Fields{}
	for k, v := range l.fields {
		next[k] = v
	}
	for k, v := range fields {
		next[k] = v
	}
	return Logger{ctx: l.ctx, fields: next}
}

// Emit writes a structured runtime log through the standard slog backend.
func (l Logger) Emit(level, message string, fields Fields) {
	ctx := l.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	merged := Fields{}
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	if _, ok := merged[FieldStatus]; !ok {
		merged[FieldStatus] = "succeeded"
	}
	if _, ok := merged["ts"]; !ok {
		merged["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	attrs := make([]slog.Attr, 0, len(merged))
	for k, v := range merged {
		attrs = append(attrs, slog.Any(k, v))
	}
	slog.LogAttrs(ctx, parseSlogLevel(level), strings.TrimSpace(message), attrs...)
}

func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
