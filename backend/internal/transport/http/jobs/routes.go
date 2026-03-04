package jobs

import (
	"errors"
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

	handler := &Handler{provider: &taskcenter.ChainStatusProvider{Providers: []taskcenter.StatusProvider{
		taskcenter.NewFrameworkStatusProvider("", ""),
		&taskcenter.LocalStatusProvider{Store: taskcenter.DefaultStore()},
	}}}
	group := rg.Group("/jobs", httpmw.EnsureTenant())
	{
		group.GET("/:taskId", handler.GetJobStatus)
	}
}

// Handler resolves job status lookups.
type Handler struct {
	provider taskcenter.StatusProvider
}

// GetJobStatus returns a job snapshot when accessible to the tenant.
func (h *Handler) GetJobStatus(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("taskId"))
	if taskID == "" {
		contracts.ResponseBadRequest(c, "task id is required")
		return
	}
	if h == nil || h.provider == nil {
		contracts.ResponseError(c, http.StatusBadGateway, "TASK_STATUS_UNAVAILABLE", "任务状态暂不可用")
		return
	}

	tenantUUID := ""
	if tenant, ok := authx.TenantUUIDFromContext(c.Request.Context()); ok {
		tenantUUID = tenant
	}

	job, err := h.provider.Get(c.Request.Context(), taskID, tenantUUID)
	if err != nil {
		if errors.Is(err, taskcenter.ErrTaskNotFound) {
			contracts.ResponseNotFound(c, "任务不存在")
			return
		}
		contracts.ResponseError(c, http.StatusBadGateway, "TASK_STATUS_UNAVAILABLE", "任务状态暂不可用")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    job,
	})
}
