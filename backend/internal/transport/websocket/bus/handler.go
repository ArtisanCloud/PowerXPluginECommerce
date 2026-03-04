package bus

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub        *Hub
	authorizer Authorizer
}

func NewHandler() *Handler {
	return &Handler{
		hub:        DefaultHub,
		authorizer: NewDefaultAuthorizer(),
	}
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) ServeWS(c *gin.Context) {
	logger.WithFields(logger.Fields{
		"component": "ws-bus",
		"path":      c.Request.URL.Path,
		"query":     c.Request.URL.RawQuery,
	}).Info("ws handshake received")

	tenantCtx, ok := authx.GetTenantContext(c)
	if !ok || tenantCtx.TenantUUID == "" {
		logger.WithFields(logger.Fields{
			"component": "ws-bus",
			"path":      c.Request.URL.Path,
		}).Warn("ws handshake rejected: tenant context missing")
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "tenant_uuid required")
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.WithError(err).WithFields(logger.Fields{
			"component":   "ws-bus",
			"tenant_uuid": tenantCtx.TenantUUID,
			"path":        c.Request.URL.Path,
		}).Warn("ws upgrade failed")
		return
	}

	client := NewClient(c.Request.Context(), conn, h.hub, h.authorizer)
	client.TenantUUID = tenantCtx.TenantUUID
	if tenantCtx.UserID > 0 {
		client.UserID = uint64(tenantCtx.UserID)
	}
	client.IsRoot = false

	h.hub.Register(client)
	_ = sendWelcome(client)
	logger.WithFields(logger.Fields{
		"component":   "ws-bus",
		"tenant_uuid": client.TenantUUID,
		"client_id":   client.ID,
	}).Info("ws client connected")
	client.Run()
	logger.WithFields(logger.Fields{
		"component":   "ws-bus",
		"tenant_uuid": client.TenantUUID,
		"client_id":   client.ID,
	}).Info("ws client disconnected")
}

func sendWelcome(client *Client) error {
	if client == nil {
		return nil
	}
	env, err := NewEnvelope(MsgTypeWelcome, "", WelcomePayload{
		Protocol:     "px.ws.bus.v1",
		Server:       "powerx-plugin-ws-bus",
		HeartbeatSec: 25,
	}, "")
	if err != nil {
		return err
	}
	client.sendEnvelope(env)
	return nil
}
