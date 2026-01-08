package pricing

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
)

type PricebooksHandler struct {
	domain *pricingsvc.Service
	pb     *pricingsvc.PricebookService
}

func NewPricebooksHandler(domain *pricingsvc.Service) *PricebooksHandler {
	if domain == nil || !domain.Ready() {
		return &PricebooksHandler{domain: domain, pb: nil}
	}
	return &PricebooksHandler{domain: domain, pb: pricingsvc.NewPricebookService(domain.Deps())}
}

func (h *PricebooksHandler) List(c *gin.Context) {
	if h == nil || h.pb == nil || !h.pb.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}

	page := intFromQuery(c, "page", 1)
	pageSize := intFromQuery(c, "page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	result, err := h.pb.List(c.Request.Context(), pricingsvc.PricebookListFilters{
		Keyword:  c.Query("keyword"),
		Type:     c.Query("type"),
		Currency: c.Query("currency"),
		Status:   c.Query("status"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	items := make([]PricebookDTO, 0, len(result.Items))
	for _, row := range result.Items {
		items = append(items, toPricebookDTO(row, nil))
	}
	c.JSON(http.StatusOK, PricebookListResponse{
		Items: items,
		Meta:  PageMeta{Page: result.Page, PageSize: result.PageSize, Total: result.Total},
	})
}

func (h *PricebooksHandler) Create(c *gin.Context) {
	if h == nil || h.pb == nil || !h.pb.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	var req PricebookCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	in := pricingsvc.CreatePricebookInput{
		Code:        req.Code,
		Name:        req.Name,
		Type:        req.Type,
		Currency:    req.Currency,
		Description: req.Description,
		Actor:       actorFromContext(c),
	}
	if req.Scopes != nil {
		in.Scopes = &pricingsvc.PricebookScopesInput{
			ChannelIDs:       req.Scopes.ChannelIDs,
			CustomerGroupIDs: req.Scopes.CustomerGroupIDs,
			SupplierIDs:      req.Scopes.SupplierIDs,
		}
	}

	pb, err := h.pb.Create(c.Request.Context(), in)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPricebookDTO(pb, req.Scopes))
}

func (h *PricebooksHandler) Update(c *gin.Context) {
	if h == nil || h.pb == nil || !h.pb.Ready() {
		respondError(c, http.StatusServiceUnavailable, pricingsvc.CodeServiceUnavailable, pricingsvc.ErrServiceUnavailable)
		return
	}
	pricebookID := strings.TrimSpace(c.Param("pricebookId"))
	if pricebookID == "" {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, errors.New("pricebookId is required"))
		return
	}
	var req PricebookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, pricingsvc.CodeInvalidArgument, err)
		return
	}

	in := pricingsvc.UpdatePricebookInput{
		PricebookID: pricebookID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Actor:       actorFromContext(c),
	}
	if req.Scopes != nil {
		in.Scopes = &pricingsvc.PricebookScopesInput{
			ChannelIDs:       req.Scopes.ChannelIDs,
			CustomerGroupIDs: req.Scopes.CustomerGroupIDs,
			SupplierIDs:      req.Scopes.SupplierIDs,
		}
	}

	pb, err := h.pb.Update(c.Request.Context(), in)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPricebookDTO(pb, req.Scopes))
}

func toPricebookDTO(pb *pricingModel.Pricebook, scopes *PricebookScopes) PricebookDTO {
	if pb == nil {
		return PricebookDTO{}
	}
	createdAt := pb.CreatedAt
	updatedAt := pb.UpdatedAt
	dto := PricebookDTO{
		ID:               pb.ID,
		Code:             pb.Code,
		Name:             pb.Name,
		Type:             pb.Type,
		Currency:         pb.Currency,
		Status:           pb.Status,
		Description:      pb.Description,
		CurrentVersionID: pb.CurrentVersionID,
		CreatedAt:        &createdAt,
		UpdatedAt:        &updatedAt,
	}
	if scopes != nil {
		dto.Scopes = &PricebookScopes{
			ChannelIDs:       scopes.ChannelIDs,
			CustomerGroupIDs: scopes.CustomerGroupIDs,
			SupplierIDs:      scopes.SupplierIDs,
		}
	}
	return dto
}

func actorFromContext(c *gin.Context) string {
	if c == nil {
		return "admin"
	}
	ctx := c.Request.Context()
	if tc, ok := authx.TenantContextFromContext(ctx); ok && tc.UserID > 0 {
		return strconv.FormatInt(tc.UserID, 10)
	}
	if tenantUUID, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tenantUUID) != "" {
		return "tenant:" + strings.TrimSpace(tenantUUID)
	}
	return "admin"
}

func intFromQuery(c *gin.Context, key string, def int) int {
	if c == nil {
		return def
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}
