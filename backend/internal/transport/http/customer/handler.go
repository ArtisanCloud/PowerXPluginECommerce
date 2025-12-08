package customer

import (
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	srv "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// Handler 负责客户域 API。
type Handler struct {
	service *srv.Service
}

// NewHandler 构造 handler。
func NewHandler(deps *app.Deps) *Handler {
	return &Handler{service: srv.NewService(deps)}
}

// ListCustomers returns mocked customer directory data for the current tenant.
func (h *Handler) ListCustomers(c *gin.Context) {
	tenantUUID, ok := httpmw.TenantUUIDString(c)
	if !ok || strings.TrimSpace(tenantUUID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context missing"})
		return
	}

	var query listQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		contracts.ResponseBadRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	filters := srv.ListFilters{
		Keyword:   query.Keyword,
		Tier:      query.Tier,
		Source:    query.Source,
		Region:    query.Region,
		RiskLevel: query.RiskLevel,
		Type:      query.Type,
		Tags:      query.normalizedTags(),
		Page:      query.effectivePage(),
		PageSize:  query.effectivePageSize(),
		Sort:      query.Sort,
	}

	result, err := h.service.ListCustomers(c.Request.Context(), tenantUUID, filters)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, result)
}

type listQuery struct {
	Keyword   string   `form:"keyword"`
	Tier      string   `form:"tier"`
	Source    string   `form:"source"`
	Region    string   `form:"region"`
	RiskLevel string   `form:"riskLevel"`
	Type      string   `form:"type"`
	Tags      []string `form:"tags"`
	Page      int      `form:"page"`
	PageSize  int      `form:"pageSize"`
	Sort      string   `form:"sort"`
}

func (q *listQuery) normalizedTags() []string {
	if len(q.Tags) == 0 {
		return nil
	}
	var tags []string
	for _, value := range q.Tags {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if strings.Contains(value, ",") {
			for _, part := range strings.Split(value, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					tags = append(tags, part)
				}
			}
			continue
		}
		tags = append(tags, strings.TrimSpace(value))
	}
	return tags
}

func (q *listQuery) effectivePage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

func (q *listQuery) effectivePageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 200 {
		return 200
	}
	return q.PageSize
}
