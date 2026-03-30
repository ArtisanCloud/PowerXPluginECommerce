package bus

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	readLimit = 1024 * 1024
)

type Client struct {
	ID         string
	TenantUUID string
	UserID     uint64
	IsRoot     bool

	ctx        context.Context
	conn       *websocket.Conn
	hub        *Hub
	authorizer Authorizer
	send       chan Envelope

	mu     sync.RWMutex
	topics map[string]struct{}
}

func NewClient(ctx context.Context, conn *websocket.Conn, hub *Hub, authorizer Authorizer) *Client {
	return &Client{
		ID:         uuid.NewString(),
		ctx:        ctx,
		conn:       conn,
		hub:        hub,
		authorizer: authorizer,
		send:       make(chan Envelope, 16),
		topics:     make(map[string]struct{}),
	}
}

func (c *Client) Run() {
	if c.conn == nil || c.hub == nil {
		return
	}
	c.conn.SetReadLimit(readLimit)
	go c.writeLoop()
	c.readLoop()
}

func (c *Client) Close() {
	if c.hub != nil {
		c.hub.Unregister(c)
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.mu.Lock()
	if c.send != nil {
		close(c.send)
		c.send = nil
	}
	c.mu.Unlock()
}

func (c *Client) readLoop() {
	defer c.Close()
	for {
		var cmd Command
		if err := c.conn.ReadJSON(&cmd); err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"component":   "ws-bus",
				"client_id":   c.ID,
				"tenant_uuid": c.TenantUUID,
			}).Info("ws read loop closed")
			return
		}
		switch strings.TrimSpace(cmd.Type) {
		case CmdSubscribe:
			c.handleSubscribe(cmd)
		case CmdUnsubscribe:
			c.handleUnsubscribe(cmd)
		case CmdPing:
			c.sendAck(cmd.ReqID, "pong", nil)
		default:
			c.sendError(cmd.ReqID, "unsupported_command", "unsupported command", "")
		}
	}
}

func (c *Client) writeLoop() {
	for env := range c.send {
		if c.conn == nil {
			return
		}
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteJSON(env); err != nil {
			logger.WithError(err).WithFields(logger.Fields{
				"component":   "ws-bus",
				"client_id":   c.ID,
				"tenant_uuid": c.TenantUUID,
			}).Info("ws write loop closed")
			return
		}
	}
}

func (c *Client) handleSubscribe(cmd Command) {
	topics := normalizeTopics(cmd)
	if len(topics) == 0 {
		c.sendError(cmd.ReqID, "bad_request", "topics required", "")
		return
	}
	allowed := make([]string, 0, len(topics))
	for _, topic := range topics {
		if c.authorizer != nil {
			if err := c.authorizer.Authorize(c.ctx, c, topic); err != nil {
				c.sendError(cmd.ReqID, "permission_denied", "subscription rejected", err.Error())
				logger.WithError(err).WithFields(logger.Fields{
					"component":   "ws-bus",
					"client_id":   c.ID,
					"tenant_uuid": c.TenantUUID,
					"topic":       topic,
				}).Warn("ws subscribe rejected")
				continue
			}
		}
		c.hub.Subscribe(c, topic)
		allowed = append(allowed, topic)
	}
	if len(allowed) == 0 {
		return
	}
	logger.WithFields(logger.Fields{
		"component":   "ws-bus",
		"client_id":   c.ID,
		"tenant_uuid": c.TenantUUID,
		"topics":      allowed,
	}).Info("ws subscribe accepted")
	c.sendAck(cmd.ReqID, "subscribed", allowed)
}

func (c *Client) handleUnsubscribe(cmd Command) {
	topics := normalizeTopics(cmd)
	if len(topics) == 0 {
		c.sendError(cmd.ReqID, "bad_request", "topics required", "")
		return
	}
	for _, topic := range topics {
		c.hub.Unsubscribe(c, topic)
	}
	c.sendAck(cmd.ReqID, "unsubscribed", topics)
}

func (c *Client) sendAck(reqID, message string, topics []string) {
	env, err := NewEnvelope(MsgTypeAck, "", AckPayload{
		ReqID:   reqID,
		OK:      true,
		Message: message,
		Topics:  topics,
	}, "")
	if err != nil {
		return
	}
	c.sendEnvelope(env)
}

func (c *Client) sendError(reqID, code, message, detail string) {
	env, err := NewEnvelope(MsgTypeError, "", ErrorPayload{
		ReqID:   reqID,
		Code:    code,
		Message: message,
		Detail:  detail,
	}, "")
	if err != nil {
		return
	}
	c.sendEnvelope(env)
}

func (c *Client) sendEnvelope(env Envelope) {
	c.mu.RLock()
	ch := c.send
	c.mu.RUnlock()
	if ch == nil {
		return
	}
	select {
	case ch <- env:
	default:
	}
}

func (c *Client) addTopic(topic string) {
	c.mu.Lock()
	c.topics[topic] = struct{}{}
	c.mu.Unlock()
}

func (c *Client) removeTopic(topic string) {
	c.mu.Lock()
	delete(c.topics, topic)
	c.mu.Unlock()
}

func normalizeTopics(cmd Command) []string {
	topics := cmd.Topics
	if cmd.Topic != "" {
		topics = append(topics, cmd.Topic)
	}
	out := make([]string, 0, len(topics))
	seen := map[string]struct{}{}
	for _, topic := range topics {
		trimmed := strings.TrimSpace(topic)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
