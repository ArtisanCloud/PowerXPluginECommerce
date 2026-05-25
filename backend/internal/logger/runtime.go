package logger

import (
	"strings"

	runtimelogging "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/logging"
	"github.com/sirupsen/logrus"
)

var runtimeFieldMasker func(Fields) Fields

// RegisterRuntimeMasker sets a hook that can rewrite runtime log fields before
// they are emitted, enabling downstream privacy masking.
func RegisterRuntimeMasker(masker func(Fields) Fields) {
	runtimeFieldMasker = masker
}

// WithRuntimeFields enriches the log entry with standard runtime metadata.
func WithRuntimeFields(pluginID, tenantID, traceID, component string, extra Fields) *logrus.Entry {
	tenantID = strings.TrimSpace(tenantID)
	traceID = strings.TrimSpace(traceID)
	fields := Fields{
		runtimelogging.FieldPluginID:   strings.TrimSpace(pluginID),
		runtimelogging.FieldTenantUUID: tenantID,
		runtimelogging.FieldTraceID:    traceID,
		runtimelogging.FieldRequestID:  traceID,
		runtimelogging.FieldComponent:  strings.TrimSpace(component),
	}
	if tenantID != "" {
		fields[runtimelogging.FieldTenantKey] = tenantID
	}
	for k, v := range extra {
		fields[k] = v
	}
	if runtimeFieldMasker != nil {
		fields = runtimeFieldMasker(fields)
	}
	return WithFields(fields)
}
