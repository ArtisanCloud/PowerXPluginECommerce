package bus

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrTopicNotAllowed  = errors.New("topic not allowed")
	ErrTenantRequired   = errors.New("tenant required")
	ErrPermissionDenied = errors.New("permission denied")
)

type Authorizer interface {
	Authorize(ctx context.Context, client *Client, topic string) error
}

type DefaultAuthorizer struct{}

func NewDefaultAuthorizer() *DefaultAuthorizer {
	return &DefaultAuthorizer{}
}

func (a *DefaultAuthorizer) Authorize(_ context.Context, client *Client, topic string) error {
	if client == nil {
		return ErrPermissionDenied
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ErrTopicNotAllowed
	}
	if client.TenantUUID == "" {
		return ErrTenantRequired
	}
	if strings.HasPrefix(topic, "_topic.") || strings.Contains(topic, "._topic.") {
		if !DefaultTopicRegistry.Exists(topic) {
			return ErrTopicNotAllowed
		}
		if !DefaultACLRegistry.Allowed(client.TenantUUID, topic, "subscribe") {
			return ErrPermissionDenied
		}
		return nil
	}
	switch topic {
	case "task.progress", "powerx.task.progress.v1", "worker.task.updated", "org_sync.progress", "powerx.org_sync.progress.v1":
		return nil
	default:
		return ErrTopicNotAllowed
	}
}
