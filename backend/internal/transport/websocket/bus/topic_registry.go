package bus

import "sync"

type TopicRegistry struct {
	mu     sync.RWMutex
	topics map[string]struct{}
}

func NewTopicRegistry() *TopicRegistry {
	return &TopicRegistry{
		topics: map[string]struct{}{},
	}
}

var DefaultTopicRegistry = NewTopicRegistry()

func (r *TopicRegistry) Register(topics []string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, topic := range topics {
		if topic == "" {
			continue
		}
		r.topics[topic] = struct{}{}
	}
}

func (r *TopicRegistry) Exists(topic string) bool {
	if r == nil || topic == "" {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.topics[topic]
	return ok
}
