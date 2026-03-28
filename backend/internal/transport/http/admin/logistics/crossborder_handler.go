package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCrossborderDocuments(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.crossborderSvc.ListDocuments(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("waybill_no")), limit)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertCrossborderDocument(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	var payload upsertCrossborderDocumentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.crossborderSvc.UpsertDocument(c.Request.Context(), tenantUUID, logisticssvc.UpsertCrossborderDocumentRequest{
		ID:          strings.TrimSpace(payload.ID),
		WaybillID:   strings.TrimSpace(payload.WaybillID),
		WaybillNo:   strings.TrimSpace(payload.WaybillNo),
		DocType:     strings.TrimSpace(payload.DocType),
		DocNo:       strings.TrimSpace(payload.DocNo),
		CountryFrom: strings.TrimSpace(payload.CountryFrom),
		CountryTo:   strings.TrimSpace(payload.CountryTo),
		Status:      strings.TrimSpace(payload.Status),
		Metadata:    payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) QuoteCrossborderTax(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	var payload quoteCrossborderTaxRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, idem, err := h.crossborderSvc.QuoteTax(c.Request.Context(), tenantUUID, logisticssvc.QuoteCrossborderTaxRequest{
		RequestKey:    strings.TrimSpace(payload.RequestKey),
		WaybillID:     strings.TrimSpace(payload.WaybillID),
		WaybillNo:     strings.TrimSpace(payload.WaybillNo),
		Destination:   strings.TrimSpace(payload.Destination),
		Currency:      strings.TrimSpace(payload.Currency),
		DeclaredValue: payload.DeclaredValue,
		ShippingFee:   payload.ShippingFee,
		InsuranceFee:  payload.InsuranceFee,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"quote":              row,
		"idempotency_status": idem,
	})
}

func (h *Handler) ListCrossborderTrackingMaps(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	var enabled *bool
	switch strings.ToLower(strings.TrimSpace(c.Query("enabled"))) {
	case "true":
		v := true
		enabled = &v
	case "false":
		v := false
		enabled = &v
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.crossborderSvc.ListTrackingMaps(c.Request.Context(), tenantUUID, strings.TrimSpace(c.Query("provider")), enabled, limit)
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) UpsertCrossborderTrackingMap(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	var payload upsertCrossborderTrackingMapRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = strings.TrimSpace(c.Param("id"))
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.crossborderSvc.UpsertTrackingMap(c.Request.Context(), tenantUUID, logisticssvc.UpsertCrossborderTrackingMapRequest{
		ID:               strings.TrimSpace(payload.ID),
		Provider:         strings.TrimSpace(payload.Provider),
		ProviderStatus:   strings.TrimSpace(payload.ProviderStatus),
		NormalizedStatus: strings.TrimSpace(payload.NormalizedStatus),
		Description:      strings.TrimSpace(payload.Description),
		Priority:         payload.Priority,
		Enabled:          payload.Enabled,
		Metadata:         payload.Metadata,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) NormalizeCrossborderTracking(c *gin.Context) {
	if h == nil || h.crossborderSvc == nil {
		contracts.ResponseServiceUnavailable(c, "crossborder service unavailable", nil)
		return
	}
	var payload normalizeCrossborderTrackingRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	normalized, source, err := h.crossborderSvc.NormalizeTrackingStatus(c.Request.Context(), tenantUUID, logisticssvc.NormalizeCrossborderTrackingRequest{
		Provider:       strings.TrimSpace(payload.Provider),
		ProviderStatus: strings.TrimSpace(payload.ProviderStatus),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{
		"normalized_status": normalized,
		"source":            source,
	})
}
