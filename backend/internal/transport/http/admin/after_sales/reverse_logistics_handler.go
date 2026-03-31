package after_sales

import (
	"errors"
	"io"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	adminsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type reverseLogisticsPayload struct {
	ReverseWaybillID string `json:"reverseWaybillId,omitempty"`
	ReverseWaybillNo string `json:"reverseWaybillNo,omitempty"`
	CarrierCode      string `json:"carrierCode,omitempty"`
	Note             string `json:"note,omitempty"`
}

type ReverseLogisticsHandler struct {
	service *adminsvc.ReverseLinkService
}

func NewReverseLogisticsHandler(service *adminsvc.ReverseLinkService) *ReverseLogisticsHandler {
	return &ReverseLogisticsHandler{service: service}
}

func (h *ReverseLogisticsHandler) LinkCase(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "after-sales reverse-link service unavailable", nil)
		return
	}
	var payload reverseLogisticsPayload
	if err := c.ShouldBindJSON(&payload); err != nil && !errors.Is(err, io.EOF) {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	caseID := strings.TrimSpace(c.Param("id"))
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	operatorID, ok := adminUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.Link(c.Request.Context(), tenantUUID, caseID, operatorID, adminsvc.ReverseLogisticsLinkInput{
		ReverseWaybillID: payload.ReverseWaybillID,
		ReverseWaybillNo: payload.ReverseWaybillNo,
		CarrierCode:      payload.CarrierCode,
		Note:             payload.Note,
	})
	if err != nil {
		respondAdminError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}
