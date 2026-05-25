package runtime_ops

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	runtimelogging "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/logging"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type probeSink struct {
	name runtimelogging.SinkType
}

func (s *probeSink) Name() runtimelogging.SinkType                        { return s.name }
func (s *probeSink) Emit(_ context.Context, _ runtimelogging.Event) error { return nil }

type probeRequest struct {
	Message    string         `json:"message"`
	Level      string         `json:"level"`
	Component  string         `json:"component"`
	TraceID    string         `json:"trace_id"`
	TenantUUID string         `json:"tenant_uuid"`
	Event      string         `json:"event"`
	Fields     map[string]any `json:"fields"`
}

func LoggingProbeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req probeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			contracts.ResponseBadRequest(c, "invalid payload")
			return
		}

		tenantUUID, mismatch := resolvePolicyTenant(c, req.TenantUUID)
		if mismatch {
			contracts.ResponseError(c, http.StatusForbidden, contracts.ErrCodeTenantMismatch, "tenant mismatch")
			return
		}
		policy := currentLoggingPolicyForTenant(tenantUUID)

		registry := runtimelogging.NewSinkRegistry()
		_ = registry.Register(&probeSink{name: runtimelogging.SinkStdout})
		_ = registry.Register(&probeSink{name: runtimelogging.SinkFile})
		_ = registry.Register(&probeSink{name: runtimelogging.SinkLoki})
		router, err := runtimelogging.NewRouter(policy, registry)
		if err != nil {
			contracts.ResponseBadRequest(c, err.Error())
			return
		}

		traceID := strings.TrimSpace(req.TraceID)
		if traceID == "" {
			traceID = strings.TrimSpace(c.GetString("request_id"))
		}

		fields := runtimelogging.Fields(req.Fields)
		if fields == nil {
			fields = runtimelogging.Fields{}
		}
		requestID := strings.TrimSpace(c.GetString("request_id"))
		fields[runtimelogging.FieldTraceID] = traceID
		fields[runtimelogging.FieldRequestID] = requestID
		fields[runtimelogging.FieldTenantUUID] = tenantUUID
		fields[runtimelogging.FieldTenantKey] = tenantUUID
		fields[runtimelogging.FieldPluginID] = app.PluginID
		component := strings.TrimSpace(req.Component)
		if component == "" {
			component = "admin.runtime.logging.probe"
		}
		fields[runtimelogging.FieldComponent] = component
		if _, ok := fields[runtimelogging.FieldStatus]; !ok {
			fields[runtimelogging.FieldStatus] = "succeeded"
		}
		if req.Event != "" {
			fields["event"] = req.Event
		}
		logger.WithRuntimeFields(app.PluginID, tenantUUID, traceID, component, logger.Fields(fields)).
			WithField("outcome_probe", true).
			Info(strings.TrimSpace(req.Message))

		outcomes := router.Route(c.Request.Context(), runtimelogging.Event{
			Message:   strings.TrimSpace(req.Message),
			Level:     strings.TrimSpace(req.Level),
			Fields:    fields,
			Timestamp: time.Now().UTC(),
		})

		loggingPolicySuccess(c, gin.H{"trace_id": traceID, "outcomes": outcomes})
	}
}
