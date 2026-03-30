package bus

import (
	"strings"
	"sync"
)

type ACLRegistry struct {
	mu      sync.RWMutex
	records map[string]map[string]map[string]struct{}
}

func NewACLRegistry() *ACLRegistry {
	return &ACLRegistry{
		records: map[string]map[string]map[string]struct{}{},
	}
}

var DefaultACLRegistry = NewACLRegistry()

func (r *ACLRegistry) Grant(tenantUUID, topic string, actions []string) {
	if r == nil {
		return
	}
	tenant := strings.TrimSpace(tenantUUID)
	topic = strings.TrimSpace(topic)
	if tenant == "" || topic == "" {
		return
	}
	normalized := normalizeGrantActions(actions)
	if len(normalized) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.records[tenant]; !ok {
		r.records[tenant] = map[string]map[string]struct{}{}
	}
	if _, ok := r.records[tenant][topic]; !ok {
		r.records[tenant][topic] = map[string]struct{}{}
	}
	for _, action := range normalized {
		r.records[tenant][topic][action] = struct{}{}
	}
}

func (r *ACLRegistry) Allowed(tenantUUID, topic, action string) bool {
	if r == nil {
		return false
	}
	tenant := strings.TrimSpace(tenantUUID)
	topic = strings.TrimSpace(topic)
	action = strings.ToLower(strings.TrimSpace(action))
	if tenant == "" || topic == "" || action == "" {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	topics, ok := r.records[tenant]
	if !ok {
		return false
	}
	actions, ok := topics[topic]
	if !ok {
		return false
	}
	_, allowed := actions[action]
	return allowed
}

func normalizeGrantActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"publish", "subscribe"}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		v := strings.ToLower(strings.TrimSpace(action))
		if v != "publish" && v != "subscribe" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
