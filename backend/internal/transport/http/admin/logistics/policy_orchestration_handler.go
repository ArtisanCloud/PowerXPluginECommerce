package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPolicyOrchestrationFlows(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.policyOrchesSvc.ListFlows(c.Request.Context(), tenantUUID, logisticssvc.PolicyOrchestrationFlowQuery{
		Status: strings.TrimSpace(c.Query("status")),
		Limit:  limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertPolicyOrchestrationFlow(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	var payload upsertPolicyOrchestrationFlowRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.policyOrchesSvc.UpsertFlow(c.Request.Context(), tenantUUID, logisticssvc.UpsertPolicyOrchestrationFlowRequest{
		ID:                strings.TrimSpace(payload.ID),
		Name:              strings.TrimSpace(payload.Name),
		Priority:          payload.Priority,
		Status:            strings.TrimSpace(payload.Status),
		FlowDefinition:    payload.FlowDefinition,
		ConflictRelations: payload.ConflictRelations,
		GrayReleaseConfig: payload.GrayReleaseConfig,
		Description:       strings.TrimSpace(payload.Description),
		OperatorID:        strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListPolicyOrchestrationVersions(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	flowID := strings.TrimSpace(c.Param("id"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.policyOrchesSvc.ListVersions(c.Request.Context(), tenantUUID, logisticssvc.PolicyOrchestrationVersionQuery{
		FlowID: flowID,
		Limit:  limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) PreviewPolicyOrchestrationConflicts(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	var payload previewPolicyOrchestrationConflictsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	result, err := h.policyOrchesSvc.PreviewConflicts(c.Request.Context(), tenantUUID, logisticssvc.PolicyOrchestrationConflictPreviewRequest{
		FlowID:            strings.TrimSpace(payload.FlowID),
		Name:              strings.TrimSpace(payload.Name),
		ConflictRelations: payload.ConflictRelations,
		FlowDefinition:    payload.FlowDefinition,
		GrayReleaseConfig: payload.GrayReleaseConfig,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, result)
}

func (h *Handler) PublishPolicyOrchestrationFlow(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	var payload publishPolicyOrchestrationFlowRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	flowID := strings.TrimSpace(c.Param("id"))
	if flowID == "" {
		flowID = strings.TrimSpace(payload.FlowID)
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.policyOrchesSvc.Publish(c.Request.Context(), tenantUUID, logisticssvc.PublishPolicyOrchestrationRequest{
		FlowID:          flowID,
		RequestKey:      strings.TrimSpace(payload.RequestKey),
		Force:           payload.Force,
		ChangeSummary:   strings.TrimSpace(payload.ChangeSummary),
		GrayReleasePlan: payload.GrayReleasePlan,
		OperatorID:      strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RollbackPolicyOrchestrationFlow(c *gin.Context) {
	if h == nil || h.policyOrchesSvc == nil {
		contracts.ResponseServiceUnavailable(c, "policy orchestration service unavailable", nil)
		return
	}
	var payload rollbackPolicyOrchestrationFlowRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	flowID := strings.TrimSpace(c.Param("id"))
	if flowID == "" {
		flowID = strings.TrimSpace(payload.FlowID)
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.policyOrchesSvc.Rollback(c.Request.Context(), tenantUUID, logisticssvc.RollbackPolicyOrchestrationRequest{
		FlowID:          flowID,
		TargetVersionID: strings.TrimSpace(payload.TargetVersionID),
		Reason:          strings.TrimSpace(payload.Reason),
		OperatorID:      strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
