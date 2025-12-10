package jobs

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires /jobs endpoints for polling task status.
func RegisterRoutes(rg *gin.RouterGroup, _ *app.Deps) {
	if rg == nil {
		return
	}
	handler := &Handler{store: taskcenter.DefaultStore()}
	group := rg.Group("/jobs", httpmw.EnsureTenant())
	{
		group.GET("/:taskId", handler.GetJobStatus)
	}
}

// Handler resolves job status lookups.
type Handler struct {
	store *taskcenter.Store
}

// GetJobStatus returns a job snapshot when accessible to the tenant.
func (h *Handler) GetJobStatus(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("taskId"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	job, ok := h.store.Get(taskID)
	if !ok {
		contracts.ResponseNotFound(c, "任务不存在")
		return
	}
	if tenant, ok := authx.TenantUUIDFromContext(c.Request.Context()); ok && tenant != "" {
		if metaTenant, _ := job.Metadata["tenantUuid"].(string); metaTenant != "" && metaTenant != tenant {
			contracts.ResponseNotFound(c, "任务不存在")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    job,
	})
}
