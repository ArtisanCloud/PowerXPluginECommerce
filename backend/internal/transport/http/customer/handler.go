package customer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	srv "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// Handler 负责客户域 API。
type Handler struct {
	service *srv.Service
}

const maxImportFileSize = 5 << 20

// NewHandler 构造 handler。
func NewHandler(deps *app.Deps) *Handler {
	return &Handler{service: srv.NewService(deps)}
}

// ListCustomers returns mocked customer directory data for the current tenant.
func (h *Handler) ListCustomers(c *gin.Context) {
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

	result, err := h.service.ListCustomers(c.Request.Context(), filters)
	if err != nil {
		if errors.Is(err, authx.ErrTenantMissing) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context missing"})
			return
		}
		contracts.ResponseInternalError(c, err)
		return
	}
	contracts.ResponseSuccess(c, result)
}

// GetCustomer returns a single customer by ID.
func (h *Handler) GetCustomer(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		contracts.ResponseBadRequest(c, "customer id is required")
		return
	}
	customer, err := h.service.GetCustomer(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, authx.ErrTenantMissing):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context missing"})
		case errors.Is(err, srv.ErrCustomerNotFound):
			contracts.ResponseNotFound(c, "客户不存在或已被删除")
		default:
			contracts.ResponseInternalError(c, err)
		}
		return
	}
	contracts.ResponseSuccess(c, customer)
}

// CreateCustomer handles POST /customers requests.
func (h *Handler) CreateCustomer(c *gin.Context) {
	var payload createCustomerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	customer, err := h.service.CreateCustomer(c.Request.Context(), payload.toInput())
	if respondMutationError(c, err) {
		return
	}
	contracts.ResponseCreated(c, customer)
}

// UpdateCustomer patches an existing record.
func (h *Handler) UpdateCustomer(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		contracts.ResponseBadRequest(c, "customer id is required")
		return
	}
	var payload updateCustomerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	customer, err := h.service.UpdateCustomer(c.Request.Context(), id, payload.toInput())
	if respondMutationError(c, err) {
		return
	}
	contracts.ResponseSuccess(c, customer)
}

// DeleteCustomer removes a record after collecting a reason.
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		contracts.ResponseBadRequest(c, "customer id is required")
		return
	}
	var payload deleteCustomerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid request body: "+err.Error())
		return
	}
	deleted, err := h.service.DeleteCustomer(c.Request.Context(), id, payload.Reason)
	if respondMutationError(c, err) {
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"id": deleted.ID}, "customer deleted")
}

// CreateImportTask accepts a CSV upload and schedules background processing.
func (h *Handler) CreateImportTask(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		contracts.ResponseBadRequest(c, "import file is required")
		return
	}
	if file.Size == 0 {
		contracts.ResponseBadRequest(c, "import file cannot be empty")
		return
	}
	if file.Size > maxImportFileSize {
		contracts.ResponseBadRequest(c, "导入文件不能超过 5MB")
		return
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		contracts.ResponseBadRequest(c, "暂仅支持 CSV 模板导入")
		return
	}
	src, err := file.Open()
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	defer src.Close()
	payload, err := io.ReadAll(src)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	taskID, err := h.service.StartImportJob(c.Request.Context(), file.Filename, payload)
	if err != nil {
		if errors.Is(err, authx.ErrTenantMissing) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context missing"})
			return
		}
		contracts.ResponseBadRequest(c, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"taskId": taskID})
}

// DownloadImportTemplate streams the CSV template used for bulk import.
func (h *Handler) DownloadImportTemplate(c *gin.Context) {
	if len(customerImportTemplate) == 0 {
		contracts.ResponseInternalError(c, errors.New("import template unavailable"))
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", importTemplateFilename))
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", customerImportTemplate)
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

func respondMutationError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, authx.ErrTenantMissing):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context missing"})
	case errors.Is(err, srv.ErrCustomerNotFound):
		contracts.ResponseNotFound(c, "客户不存在或已被删除")
	case errors.Is(err, srv.ErrCustomerConflict):
		contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeConflict, conflictMessage(err))
	default:
		var validationErrs srv.ValidationErrors
		if errors.As(err, &validationErrs) {
			contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", validationErrs)
		} else {
			contracts.ResponseInternalError(c, err)
		}
	}
	return true
}

func conflictMessage(err error) string {
	msg := "客户数据已存在"
	parts := strings.Split(err.Error(), ":")
	if len(parts) > 1 {
		field := strings.TrimSpace(parts[len(parts)-1])
		if field == "email" {
			return "邮箱已占用，请更换后重试"
		}
		if field == "phone" {
			return "手机号已占用，请更换后重试"
		}
	}
	return msg
}

type createCustomerRequest struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone"`
	Source         string   `json:"source"`
	Country        string   `json:"country"`
	Region         string   `json:"region"`
	MembershipTier string   `json:"membershipTier"`
	AccountManager string   `json:"accountManager"`
	Tags           []string `json:"tags"`
	Notes          string   `json:"notes"`
}

func (r createCustomerRequest) toInput() srv.CreateCustomerInput {
	return srv.CreateCustomerInput{
		Name:           r.Name,
		Type:           r.Type,
		Email:          r.Email,
		Phone:          r.Phone,
		Source:         r.Source,
		Country:        r.Country,
		Region:         r.Region,
		MembershipTier: r.MembershipTier,
		AccountManager: r.AccountManager,
		Tags:           r.Tags,
		Notes:          r.Notes,
	}
}

type updateCustomerRequest struct {
	Name           *string   `json:"name"`
	Type           *string   `json:"type"`
	Email          *string   `json:"email"`
	Phone          *string   `json:"phone"`
	Source         *string   `json:"source"`
	Country        *string   `json:"country"`
	Region         *string   `json:"region"`
	MembershipTier *string   `json:"membershipTier"`
	AccountManager *string   `json:"accountManager"`
	Status         *string   `json:"status"`
	Tags           *[]string `json:"tags"`
	Notes          *string   `json:"notes"`
}

func (r updateCustomerRequest) toInput() srv.UpdateCustomerInput {
	return srv.UpdateCustomerInput{
		Name:           r.Name,
		Type:           r.Type,
		Email:          r.Email,
		Phone:          r.Phone,
		Source:         r.Source,
		Country:        r.Country,
		Region:         r.Region,
		MembershipTier: r.MembershipTier,
		AccountManager: r.AccountManager,
		Status:         r.Status,
		Tags:           r.Tags,
		Notes:          r.Notes,
	}
}

type deleteCustomerRequest struct {
	Reason string `json:"reason"`
}
