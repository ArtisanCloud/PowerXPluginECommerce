package master

import (
	"context"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/taskbus"
)

// TaskBusEmitter bridges channel events to TaskBus topics.
type TaskBusEmitter struct {
	bus    taskbus.Publisher
	prefix string
}

// NewTaskBusEmitter builds an emitter using global deps config.
func NewTaskBusEmitter(deps *app.Deps, bus taskbus.Publisher) *TaskBusEmitter {
	if bus == nil && deps != nil {
		bus = deps.TaskBus
	}
	prefix := "powerx.channel"
	if deps != nil && deps.Config != nil {
		prefix = deps.Config.TaskBusTopicPrefix()
	}
	return &TaskBusEmitter{bus: bus, prefix: prefix}
}

// Emit publishes structured payloads onto TaskBus when enabled.
func (e *TaskBusEmitter) Emit(ctx context.Context, topic string, payload map[string]any) {
	if e == nil || e.bus == nil {
		return
	}
	finalTopic := topic
	if e.prefix != "" {
		finalTopic = e.prefix + "." + topic
	}
	_ = e.bus.Publish(ctx, taskbus.Event{Topic: finalTopic, Payload: payload})
}
