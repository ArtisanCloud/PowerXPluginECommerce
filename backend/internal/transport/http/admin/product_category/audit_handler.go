package product_category

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	adminconsole "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/admin_console"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type auditRecord struct {
	ID             string `json:"id"`
	Action         string `json:"action"`
	PermissionCode string `json:"permissionCode"`
	ActorID        string `json:"actorId"`
	Summary        string `json:"summary,omitempty"`
	Diff           any    `json:"diff,omitempty"`
	OccurredAt     string `json:"occurredAt"`
}

func (h *Handler) CategoryAudit(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	categoryID := strings.TrimSpace(c.Param("id"))
	if categoryID == "" {
		contracts.ResponseBadRequest(c, "category id is required")
		return
	}
	limit := toInt(c.Query("limit"), 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	tenantID, err := h.service.TenantUUID(c.Request.Context())
	if err != nil {
		contracts.ResponseError(c, http.StatusUnauthorized, contracts.ErrCodeUnauthorized, err.Error())
		return
	}

	var events []adminconsole.AuditEvent
	db := h.service.DB()
	if db == nil {
		contracts.ResponseServiceUnavailable(c, "database unavailable", nil)
		return
	}
	query := db.WithContext(c.Request.Context()).
		Model(&adminconsole.AuditEvent{}).
		Where("plugin_id = ? AND tenant_uuid = ? AND resource_ref = ?", app.PluginID, tenantID, categoryID).
		Order("occurred_at DESC").
		Limit(limit)
	if err := query.Find(&events).Error; err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	items := make([]auditRecord, 0, len(events))
	for _, e := range events {
		summary := ""
		if e.Summary != nil {
			summary = *e.Summary
		}
		var diff any
		if len(e.Diff) > 0 {
			_ = json.Unmarshal(e.Diff, &diff)
		}
		items = append(items, auditRecord{
			ID:             e.ID,
			Action:         e.Action,
			PermissionCode: e.PermissionCode,
			ActorID:        e.ActorID,
			Summary:        summary,
			Diff:           diff,
			OccurredAt:     e.OccurredAt.Format(time.RFC3339),
		})
	}
	contracts.ResponseSuccess(c, gin.H{"items": items})
}
