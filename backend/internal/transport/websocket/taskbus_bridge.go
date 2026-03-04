package websocket

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/taskbus"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket/bus"
)

var bridgedTaskTopics = []string{
	"task.progress",
	"powerx.task.progress.v1",
	"worker.task.updated",
	"org_sync.progress",
	"powerx.org_sync.progress.v1",
}

func RegisterTaskBusBridge(client taskbus.Client) {
	if client == nil {
		return
	}
	for _, topic := range bridgedTaskTopics {
		topicName := strings.TrimSpace(topic)
		if topicName == "" {
			continue
		}
		err := client.Subscribe(topicName, func(ctx context.Context, evt taskbus.Event) error {
			tenantUUID := resolveBridgeTenant(ctx, evt)
			traceID := strings.TrimSpace(evt.Metadata["trace_id"])
			if traceID == "" {
				traceID = strings.TrimSpace(evt.Metadata["request_id"])
			}
			topic := strings.TrimSpace(evt.Topic)
			bus.DefaultHub.Publish(tenantUUID, topic, evt.Payload, traceID)
			logger.WithFields(logger.Fields{
				"component":   "taskbus-ws-bridge",
				"topic":       topic,
				"tenant_uuid": tenantUUID,
				"trace_id":    traceID,
				"subscribers": bus.DefaultHub.TopicSubscribers(topic),
			}).Info("taskbus event bridged to ws")
			return nil
		})
		if err != nil {
			logger.WithError(err).WithField("topic", topicName).Warn("taskbus ws bridge subscribe failed")
		}
	}
}

func resolveBridgeTenant(ctx context.Context, evt taskbus.Event) string {
	tenantUUID := strings.TrimSpace(evt.Metadata["tenant_uuid"])
	if tenantUUID != "" {
		return tenantUUID
	}
	if payloadMap, ok := evt.Payload.(map[string]any); ok {
		if value, ok := payloadMap["tenant_uuid"]; ok {
			if parsed := strings.TrimSpace(toString(value)); parsed != "" {
				return parsed
			}
		}
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok {
		if parsed := strings.TrimSpace(tid); parsed != "" {
			return parsed
		}
	}
	return ""
}

func toString(raw any) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}
