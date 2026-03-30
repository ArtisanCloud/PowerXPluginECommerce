package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCustomsRulePacks(c *gin.Context) {
	if h == nil || h.customsSvc == nil {
		contracts.ResponseServiceUnavailable(c, "customs rule service unavailable", nil)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.customsSvc.ListPacks(
		c.Request.Context(),
		tenantUUID,
		strings.TrimSpace(c.Query("country_code")),
		strings.TrimSpace(c.Query("status")),
		limit,
	)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertCustomsRulePack(c *gin.Context) {
	if h == nil || h.customsSvc == nil {
		contracts.ResponseServiceUnavailable(c, "customs rule service unavailable", nil)
		return
	}
	var payload upsertCustomsRulePackRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.customsSvc.UpsertPack(c.Request.Context(), tenantUUID, logisticssvc.UpsertCustomsRulePackRequest{
		ID:               strings.TrimSpace(payload.ID),
		Name:             strings.TrimSpace(payload.Name),
		CountryCode:      strings.TrimSpace(payload.CountryCode),
		Status:           strings.TrimSpace(payload.Status),
		Strategy:         strings.TrimSpace(payload.Strategy),
		DefaultRiskLevel: strings.TrimSpace(payload.DefaultRiskLevel),
		Description:      strings.TrimSpace(payload.Description),
		Metadata:         payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) ListCustomsRuleVersions(c *gin.Context) {
	if h == nil || h.customsSvc == nil {
		contracts.ResponseServiceUnavailable(c, "customs rule service unavailable", nil)
		return
	}
	packID := strings.TrimSpace(c.Param("id"))
	if packID == "" {
		contracts.ResponseBadRequest(c, "pack id is required")
		return
	}
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.customsSvc.ListVersions(c.Request.Context(), tenantUUID, packID, limit)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) PublishCustomsRuleVersion(c *gin.Context) {
	if h == nil || h.customsSvc == nil {
		contracts.ResponseServiceUnavailable(c, "customs rule service unavailable", nil)
		return
	}
	packID := strings.TrimSpace(c.Param("id"))
	if packID == "" {
		contracts.ResponseBadRequest(c, "pack id is required")
		return
	}
	var payload publishCustomsRuleVersionRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	rules := make([]logisticssvc.CustomsRuleDefinition, 0, len(payload.Rules))
	for _, item := range payload.Rules {
		rules = append(rules, logisticssvc.CustomsRuleDefinition{
			Code:       strings.TrimSpace(item.Code),
			Name:       strings.TrimSpace(item.Name),
			Field:      strings.TrimSpace(item.Field),
			Operator:   strings.TrimSpace(item.Operator),
			Value:      item.Value,
			RiskLevel:  strings.TrimSpace(item.RiskLevel),
			Suggestion: strings.TrimSpace(item.Suggestion),
			Enabled:    item.Enabled,
		})
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.customsSvc.PublishVersion(c.Request.Context(), tenantUUID, logisticssvc.PublishCustomsRuleVersionRequest{
		PackID:      packID,
		VersionNo:   payload.VersionNo,
		Status:      strings.TrimSpace(payload.Status),
		HitStrategy: strings.TrimSpace(payload.HitStrategy),
		Rules:       rules,
		RiskConfig:  payload.RiskConfig,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) CustomsPrecheck(c *gin.Context) {
	if h == nil || h.customsSvc == nil {
		contracts.ResponseServiceUnavailable(c, "customs rule service unavailable", nil)
		return
	}
	var payload customsPrecheckRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.customsSvc.Precheck(c.Request.Context(), tenantUUID, logisticssvc.CustomsPrecheckRequest{
		PackID:        strings.TrimSpace(payload.PackID),
		CountryCode:   strings.TrimSpace(payload.CountryCode),
		WaybillNo:     strings.TrimSpace(payload.WaybillNo),
		DeclaredValue: payload.DeclaredValue,
		TaxNo:         strings.TrimSpace(payload.TaxNo),
		HSCode:        strings.TrimSpace(payload.HSCode),
		DocumentType:  strings.TrimSpace(payload.DocumentType),
		DocumentCount: payload.DocumentCount,
		ManualRelease: payload.ManualRelease,
		Payload:       payload.Payload,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
