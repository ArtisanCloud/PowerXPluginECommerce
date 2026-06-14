package promotion

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func requireAdminUser(c *gin.Context) (string, bool) {
	actor, ok := authx.RequireActorFromGin(c)
	if !ok {
		contracts.ResponseError(c, http.StatusUnauthorized, contracts.ErrCodeUnauthorized, "unauthorized")
		return "", false
	}
	return actor.ID(), true
}

func requestIDFromRequest(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v := strings.TrimSpace(c.GetHeader("X-Request-ID")); v != "" {
		return v
	}
	return strings.TrimSpace(c.GetHeader("Request-ID"))
}

func parsePage(raw string, fallback int) int {
	v, _ := strconv.Atoi(strings.TrimSpace(raw))
	if v <= 0 {
		return fallback
	}
	return v
}
