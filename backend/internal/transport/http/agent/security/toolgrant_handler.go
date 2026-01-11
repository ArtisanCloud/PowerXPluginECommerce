package security

import (
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	toolgrantservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/tool_grant"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type ToolGrantHandler struct {
	service *toolgrantservice.Service
}

func NewToolGrantHandler(deps *app.Deps) *ToolGrantHandler {
	if deps == nil || deps.DB == nil || deps.Config == nil || deps.Config.Security == nil || deps.Config.Security.ToolGrantSecret == "" {
		return &ToolGrantHandler{service: nil}
	}
	signingKey := []byte(deps.Config.Security.ToolGrantSecret)
	logger := deps.RuntimeLogger(deps.Ctx, "agent_toolgrant", nil)
	svc := toolgrantservice.NewService(deps.DB, deps.Config, logger, signingKey)
	return &ToolGrantHandler{service: svc}
}

func (h *ToolGrantHandler) Verify(c *gin.Context) {
	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "toolgrant is not configured"})
		return
	}
	var payload struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	tenantID, ok := middleware.TenantUUIDFromContext(c.Request.Context())
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant context missing"})
		return
	}
	claims, err := h.service.Validate(c.Request.Context(), tenantID, payload.Token)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "claims": claims})
}
