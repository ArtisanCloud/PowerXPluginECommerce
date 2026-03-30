package reverse

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	domainName = "reverse"
	eventMsg   = "reverse_event"
	auditMsg   = "reverse_audit_event"
)

// Emitter emits structured reverse-logistics observability events.
type Emitter struct {
	entry *logrus.Entry
}

// Event describes a domain event emitted by reverse services.
type Event struct {
	Action      string
	TenantUUID  string
	RequestID   string
	WaybillNo   string
	AfterSaleID string
	Status      string
	Result      string
	Metadata    map[string]any
	EmittedAt   time.Time
}

// AuditEvent captures audit-oriented fields.
type AuditEvent struct {
	Action     string
	TenantUUID string
	ActorID    string
	TargetID   string
	Result     string
	Reason     string
	Metadata   map[string]any
	EmittedAt  time.Time
}

// NewEmitter constructs a reverse-logistics event emitter.
func NewEmitter(entry *logrus.Entry) *Emitter {
	if entry == nil {
		entry = logrus.New().WithField("component", "reverse-observability")
	}
	return &Emitter{entry: entry}
}

func (e *Emitter) Emit(evt Event) {
	if e == nil || e.entry == nil {
		return
	}
	action := strings.TrimSpace(evt.Action)
	if action == "" {
		return
	}
	fields := logrus.Fields{
		"domain":     domainName,
		"action":     action,
		"emitted_at": timestamp(evt.EmittedAt),
	}
	add(fields, "tenant_uuid", evt.TenantUUID)
	add(fields, "request_id", evt.RequestID)
	add(fields, "waybill_no", evt.WaybillNo)
	add(fields, "after_sale_id", evt.AfterSaleID)
	add(fields, "status", evt.Status)
	add(fields, "result", evt.Result)
	merge(fields, evt.Metadata)
	e.entry.WithFields(fields).Info(eventMsg)
}

func (e *Emitter) EmitAudit(evt AuditEvent) {
	if e == nil || e.entry == nil {
		return
	}
	action := strings.TrimSpace(evt.Action)
	if action == "" {
		return
	}
	fields := logrus.Fields{
		"domain":     domainName,
		"action":     action,
		"audit":      true,
		"emitted_at": timestamp(evt.EmittedAt),
	}
	add(fields, "tenant_uuid", evt.TenantUUID)
	add(fields, "actor_id", evt.ActorID)
	add(fields, "target_id", evt.TargetID)
	add(fields, "result", evt.Result)
	add(fields, "reason", evt.Reason)
	merge(fields, evt.Metadata)
	e.entry.WithFields(fields).Info(auditMsg)
}

func timestamp(ts time.Time) string {
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return ts.UTC().Format(time.RFC3339Nano)
}

func add(fields logrus.Fields, key, value string) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
		return
	}
	fields[key] = strings.TrimSpace(value)
}

func merge(fields logrus.Fields, metadata map[string]any) {
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
