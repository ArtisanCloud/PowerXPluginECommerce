package taskbus

import (
	"context"
	"sync"

	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/sirupsen/logrus"
)

// Event represents a TaskBus message that can be published or consumed.
type Event struct {
	Topic    string
	Payload  any
	Metadata map[string]string
}

// Handler processes a TaskBus event.
type Handler func(ctx context.Context, evt Event) error

// Publisher exposes the ability to publish TaskBus events.
type Publisher interface {
	Publish(ctx context.Context, evt Event) error
}

// Subscriber registers handlers for TaskBus topics.
type Subscriber interface {
	Subscribe(topic string, handler Handler) error
}

// Client combines publishing and subscribing capabilities.
type Client interface {
	Publisher
	Subscriber
	Close(context.Context) error
}

// LocalClient is a lightweight in-memory TaskBus implementation for dev/test.
type LocalClient struct {
	logger   *logrus.Entry
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewLocalClient creates a logging TaskBus client.
func NewLocalClient(logger *logrus.Entry) *LocalClient {
	if logger == nil {
		logger = pxlogger.WithField("component", "taskbus-local")
	}
	return &LocalClient{logger: logger, handlers: map[string][]Handler{}}
}

// Publish logs the event and synchronously invokes registered handlers.
func (c *LocalClient) Publish(ctx context.Context, evt Event) error {
	if c == nil {
		return nil
	}
	c.logger.WithFields(logrus.Fields{
		"topic":    evt.Topic,
		"metadata": evt.Metadata,
		"payload":  evt.Payload,
	}).Info("taskbus.publish")

	c.mu.RLock()
	handlers := append([]Handler{}, c.handlers[evt.Topic]...)
	c.mu.RUnlock()

	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if err := handler(ctx, evt); err != nil {
			c.logger.WithError(err).WithField("topic", evt.Topic).Warn("taskbus handler error")
		}
	}
	return nil
}

// Subscribe adds a handler for the specified topic.
func (c *LocalClient) Subscribe(topic string, handler Handler) error {
	if c == nil || handler == nil || topic == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = append(c.handlers[topic], handler)
	return nil
}

// Close releases resources, no-op for local client.
func (c *LocalClient) Close(context.Context) error {
	return nil
}

// NoopClient discards all TaskBus interactions.
type NoopClient struct{}

// NewNoopClient returns a Publisher that drops events.
func NewNoopClient() *NoopClient { return &NoopClient{} }

// Publish implements Publisher.
func (NoopClient) Publish(context.Context, Event) error { return nil }

// Subscribe implements Subscriber.
func (NoopClient) Subscribe(string, Handler) error { return nil }

// Close implements Client.
func (NoopClient) Close(context.Context) error { return nil }
