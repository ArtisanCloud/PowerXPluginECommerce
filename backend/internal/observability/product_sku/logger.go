package product_sku

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	eventDomain       = "product_sku"
	eventMessage      = "product_sku_event"
	auditEventMessage = "product_sku_audit_event"
)

// Logger emits structured product SKU events for audit/observability sinks.
type Logger struct {
	entry *logrus.Entry
}

// Event represents the canonical structure emitted for SKU observability.
type Event struct {
	Action        string
	TenantID      string
	ActorID       string
	RequestID     string
	SPUID         string
	SKUID         string
	TaskID        string
	ChannelCode   string
	TargetType    string
	TargetID      string
	Status        string
	Result        string
	Operation     string
	Severity      string
	AffectedCount int
	Duration      time.Duration
	Error         error
	Metadata      map[string]any
	EmittedAt     time.Time
}

// AuditEvent captures an audit-friendly subset of fields that must be searchable.
type AuditEvent struct {
	Action     string
	TenantID   string
	ActorID    string
	TargetType string
	TargetID   string
	Result     string
	Reason     string
	Metadata   map[string]any
	EmittedAt  time.Time
}

// NewLogger constructs a logger with sane defaults.
func NewLogger(entry *logrus.Entry) *Logger {
	if entry == nil {
		entry = logrus.New().WithField("component", "product-sku")
	}
	return &Logger{entry: entry}
}

// Emit outputs a structured event with action + payload.
func (l *Logger) Emit(action string, payload map[string]any) {
	l.EmitEvent(Event{
		Action:   action,
		Metadata: payload,
	})
}

// EmitEvent records a structured event with canonical fields.
func (l *Logger) EmitEvent(evt Event) {
	if l == nil || l.entry == nil {
		return
	}
	action := strings.TrimSpace(evt.Action)
	if action == "" {
		return
	}
	fields := evt.asFields()
	l.entry.WithFields(fields).Info(eventMessage)
}

// EmitAudit records an audit-centric log entry to aid approval trails.
func (l *Logger) EmitAudit(evt AuditEvent) {
	if l == nil || l.entry == nil {
		return
	}
	action := strings.TrimSpace(evt.Action)
	if action == "" {
		return
	}
	fields := logrus.Fields{
		"action":     action,
		"domain":     eventDomain,
		"audit":      true,
		"emitted_at": formatTimestamp(evt.EmittedAt),
	}
	addString(fields, "tenant_id", evt.TenantID)
	addString(fields, "actor_id", evt.ActorID)
	addString(fields, "target_type", evt.TargetType)
	addString(fields, "target_id", evt.TargetID)
	addString(fields, "result", evt.Result)
	addString(fields, "reason", evt.Reason)
	mergeMetadata(fields, evt.Metadata)
	l.entry.WithFields(fields).Info(auditEventMessage)
}

// WithTenant annotates downstream events with tenant context.
func (l *Logger) WithTenant(tenantID string) *Logger {
	if l == nil || l.entry == nil {
		return l
	}
	return &Logger{entry: l.entry.WithField("tenant_id", tenantID)}
}

func (evt Event) asFields() logrus.Fields {
	fields := logrus.Fields{
		"action":     strings.TrimSpace(evt.Action),
		"domain":     eventDomain,
		"emitted_at": formatTimestamp(evt.EmittedAt),
	}
	addString(fields, "tenant_id", evt.TenantID)
	addString(fields, "actor_id", evt.ActorID)
	addString(fields, "request_id", evt.RequestID)
	addString(fields, "spu_id", evt.SPUID)
	addString(fields, "sku_id", evt.SKUID)
	addString(fields, "task_id", evt.TaskID)
	addString(fields, "channel_code", evt.ChannelCode)
	addString(fields, "target_type", evt.TargetType)
	addString(fields, "target_id", evt.TargetID)
	addString(fields, "status", evt.Status)
	addString(fields, "result", evt.Result)
	addString(fields, "operation", evt.Operation)
	addString(fields, "severity", evt.Severity)
	if evt.AffectedCount > 0 {
		fields["affected_count"] = evt.AffectedCount
	}
	if evt.Duration > 0 {
		fields["duration_ms"] = evt.Duration.Milliseconds()
	}
	if evt.Error != nil {
		fields["error"] = evt.Error.Error()
	}
	mergeMetadata(fields, evt.Metadata)
	return fields
}

func mergeMetadata(fields logrus.Fields, metadata map[string]any) {
	if len(metadata) == 0 {
		return
	}
	for k, v := range metadata {
		if k == "" || v == nil {
			continue
		}
		if _, exists := fields[k]; exists {
			continue
		}
		fields[k] = v
	}
}

func addString(fields logrus.Fields, key, value string) {
	if strings.TrimSpace(key) == "" {
		return
	}
	if strings.TrimSpace(value) == "" {
		return
	}
	fields[key] = strings.TrimSpace(value)
}

func formatTimestamp(ts time.Time) string {
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return ts.UTC().Format(time.RFC3339Nano)
}
