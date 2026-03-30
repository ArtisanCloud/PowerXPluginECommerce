package websocket

import (
	"path"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/websocket/bus"
	"github.com/gin-gonic/gin"
)

func RegisterWSRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc, cfg *config.Config) {
	if r == nil {
		return
	}

	apiPrefix := "/api/v1"
	if cfg != nil && cfg.Server != nil {
		if prefix := strings.TrimSpace(cfg.Server.APIPrefix); prefix != "" {
			apiPrefix = prefix
		}
	}
	if !strings.HasPrefix(apiPrefix, "/") {
		apiPrefix = "/" + apiPrefix
	}
	apiPrefix = strings.TrimRight(apiPrefix, "/")

	defaultPath := path.Join(apiPrefix, "ws")
	configuredPath := defaultPath
	if cfg != nil && cfg.Server != nil {
		if p := strings.TrimSpace(cfg.Server.WSPrefix); p != "" {
			configuredPath = p
		}
	}
	if !strings.HasPrefix(configuredPath, "/") {
		configuredPath = "/" + configuredPath
	}
	configuredPath = strings.TrimRight(configuredPath, "/")

	paths := []string{configuredPath, defaultPath, "/api/ws"}
	seen := map[string]struct{}{}

	wsHandler := bus.NewHandler()
	for _, wsPath := range paths {
		wsPath = strings.TrimSpace(wsPath)
		if wsPath == "" {
			continue
		}
		if !strings.HasPrefix(wsPath, "/") {
			wsPath = "/" + wsPath
		}
		wsPath = strings.TrimRight(wsPath, "/")
		if wsPath == "" {
			wsPath = "/"
		}
		if _, ok := seen[wsPath]; ok {
			continue
		}
		seen[wsPath] = struct{}{}

		wsGroup := r.Group(wsPath)
		wsGroup.Use(BearerShim())
		if authMiddleware != nil {
			wsGroup.Use(authMiddleware)
		}
		wsGroup.GET("", wsHandler.ServeWS)
	}
}
