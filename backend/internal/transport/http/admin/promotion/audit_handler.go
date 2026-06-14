package promotion

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	svc *promotionsvc.AuditLogService
}

func NewAuditHandler(svc *promotionsvc.AuditLogService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *gin.Context) {
	if h == nil || h.svc == nil || !h.svc.Ready() {
		contracts.ResponseServiceUnavailable(c, "promotion audit service unavailable", nil)
		return
	}
	tenantUUID, ok := httpmw.TenantUUIDFromContext(c)
	if !ok {
		contracts.ResponseUnauthorized(c, "tenant context missing")
		return
	}
	rows, total, err := h.svc.List(c.Request.Context(), tenantUUID, c.Param("id"), parsePage(c.Query("page"), 1), parsePage(c.Query("page_size"), 20))
	if err != nil {
		respondError(c, err)
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows, "total": total})
}
