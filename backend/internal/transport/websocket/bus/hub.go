package bus

import (
	"context"
	"strings"
	"sync"

	frameworkwsbus "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/runtime/wsbus"
)

type Hub struct {
	mu          sync.RWMutex
	sessions    map[string]*Client
	subscribers map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		sessions:    make(map[string]*Client),
		subscribers: make(map[string]map[string]*Client),
	}
}

var DefaultHub = NewHub()
var DefaultPublisher frameworkwsbus.Publisher = frameworkwsbus.NewLocalPublisher(DefaultHub, nil)

func (h *Hub) Register(client *Client) {
	if client == nil {
		return
	}
	h.mu.Lock()
	h.sessions[client.ID] = client
	h.mu.Unlock()
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}
	h.mu.Lock()
	delete(h.sessions, client.ID)
	for topic, subs := range h.subscribers {
		delete(subs, client.ID)
		if len(subs) == 0 {
			delete(h.subscribers, topic)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Subscribe(client *Client, topic string) {
	if client == nil || topic == "" {
		return
	}
	h.mu.Lock()
	if _, ok := h.subscribers[topic]; !ok {
		h.subscribers[topic] = make(map[string]*Client)
	}
	h.subscribers[topic][client.ID] = client
	h.mu.Unlock()
	client.addTopic(topic)
}

func (h *Hub) Unsubscribe(client *Client, topic string) {
	if client == nil || topic == "" {
		return
	}
	h.mu.Lock()
	if subs, ok := h.subscribers[topic]; ok {
		delete(subs, client.ID)
		if len(subs) == 0 {
			delete(h.subscribers, topic)
		}
	}
	h.mu.Unlock()
	client.removeTopic(topic)
}

func (h *Hub) Publish(_ context.Context, topic string, payload any, opts frameworkwsbus.PublishOptions) error {
	h.publishToSubscribers(strings.TrimSpace(opts.TenantUUID), strings.TrimSpace(topic), payload, strings.TrimSpace(opts.TraceID))
	return nil
}

func (h *Hub) publishToSubscribers(tenantUUID, topic string, payload any, traceID string) {
	if topic == "" {
		return
	}
	env, err := NewEnvelope(MsgTypeEvent, topic, payload, traceID)
	if err != nil {
		return
	}
	trimmedTenant := tenantUUID
	h.mu.RLock()
	subs := h.subscribers[topic]
	for _, client := range subs {
		if client == nil {
			continue
		}
		if trimmedTenant != "" && client.TenantUUID != trimmedTenant {
			continue
		}
		client.sendEnvelope(env)
	}
	h.mu.RUnlock()
}

func (h *Hub) TopicSubscribers(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers[topic])
}
