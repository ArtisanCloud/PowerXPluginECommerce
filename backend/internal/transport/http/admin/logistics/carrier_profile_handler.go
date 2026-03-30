package logistics

import (
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCarrierProfiles(c *gin.Context) {
	if h == nil || h.carrierProfileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "carrier profile service unavailable", nil)
		return
	}
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.carrierProfileSvc.List(c.Request.Context(), tenantUUID, logisticssvc.CarrierProfileQuery{
		Status: strings.TrimSpace(c.Query("status")),
		Limit:  limit,
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) EvaluateCarrierProfiles(c *gin.Context) {
	if h == nil || h.carrierProfileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "carrier profile service unavailable", nil)
		return
	}
	var payload evaluateCarrierProfilesRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	rows, err := h.carrierProfileSvc.Evaluate(c.Request.Context(), tenantUUID, logisticssvc.EvaluateCarrierProfilesRequest{
		CarrierID:  strings.TrimSpace(payload.CarrierID),
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": rows})
}

func (h *Handler) ConfirmCarrierProfileRating(c *gin.Context) {
	if h == nil || h.carrierProfileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "carrier profile service unavailable", nil)
		return
	}
	profileID := strings.TrimSpace(c.Param("id"))
	if profileID == "" {
		contracts.ResponseBadRequest(c, "profile id is required")
		return
	}
	var payload confirmCarrierProfileRatingRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.carrierProfileSvc.ConfirmRating(c.Request.Context(), tenantUUID, profileID, logisticssvc.ConfirmCarrierProfileRatingRequest{
		Rating:     strings.TrimSpace(payload.Rating),
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RetireCarrierProfile(c *gin.Context) {
	if h == nil || h.carrierProfileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "carrier profile service unavailable", nil)
		return
	}
	profileID := strings.TrimSpace(c.Param("id"))
	if profileID == "" {
		contracts.ResponseBadRequest(c, "profile id is required")
		return
	}
	var payload retireCarrierProfileRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.carrierProfileSvc.Retire(c.Request.Context(), tenantUUID, profileID, logisticssvc.RetireCarrierProfileRequest{
		Reason:     strings.TrimSpace(payload.Reason),
		Force:      payload.Force,
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}

func (h *Handler) RestoreCarrierProfile(c *gin.Context) {
	if h == nil || h.carrierProfileSvc == nil {
		contracts.ResponseServiceUnavailable(c, "carrier profile service unavailable", nil)
		return
	}
	profileID := strings.TrimSpace(c.Param("id"))
	if profileID == "" {
		contracts.ResponseBadRequest(c, "profile id is required")
		return
	}
	var payload restoreCarrierProfileRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	row, err := h.carrierProfileSvc.Restore(c.Request.Context(), tenantUUID, profileID, logisticssvc.RestoreCarrierProfileRequest{
		OperatorID: strings.TrimSpace(payload.OperatorID),
	})
	if err != nil {
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, row)
}
