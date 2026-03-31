package after_sales

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	afterSalesSvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	caseService  *afterSalesSvc.CaseService
	queryService *afterSalesSvc.QueryService
}

func NewHandler(deps *app.Deps) *Handler {
	return &Handler{
		caseService:  afterSalesSvc.NewCaseService(deps),
		queryService: afterSalesSvc.NewQueryService(deps),
	}
}

func (h *Handler) CreateCase(c *gin.Context) {
	if h == nil || h.caseService == nil {
		respondMiniAppError(c, errors.New("after-sales handler unavailable"))
		return
	}
	var req afterSalesSvc.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondMiniAppError(c, err)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	resp, err := h.caseService.Create(c.Request.Context(), tenantUUID, cc.CustomerID, req)
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) ListCases(c *gin.Context) {
	if h == nil || h.queryService == nil {
		respondMiniAppError(c, errors.New("after-sales handler unavailable"))
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	status := strings.TrimSpace(c.Query("status"))
	resp, err := h.queryService.ListByCustomer(c.Request.Context(), tenantUUID, cc.CustomerID, status, page, pageSize)
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) GetCase(c *gin.Context) {
	if h == nil || h.queryService == nil {
		respondMiniAppError(c, errors.New("after-sales handler unavailable"))
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	cc, _ := authx.GetCustomerContext(c)
	resp, err := h.queryService.GetDetailByCustomer(c.Request.Context(), tenantUUID, cc.CustomerID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		respondMiniAppError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func respondMiniAppError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	logger.WithError(err).WithFields(logger.Fields{
		"path":   c.FullPath(),
		"method": c.Request.Method,
	}).Error("mini-app after-sales request failed")

	status, code, message := mapHTTPError(err)
	contracts.ResponseError(c, status, code, message)
}

func mapHTTPError(err error) (int, string, string) {
	switch {
	case err == nil:
		return http.StatusOK, "", ""
	case errors.Is(err, afterSalesSvc.ErrServiceUnavailable):
		return http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "售后服务暂不可用"
	case errors.Is(err, afterSalesSvc.ErrCustomerRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized, "未登录或客户信息缺失"
	case errors.Is(err, afterSalesSvc.ErrInvalidRequest), errors.Is(err, afterSalesSvc.ErrInvalidCaseType):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "请求参数不合法"
	case errors.Is(err, afterSalesSvc.ErrOrderNotFound), errors.Is(err, afterSalesSvc.ErrOrderItemNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound, "订单或订单明细不存在"
	case errors.Is(err, afterSalesSvc.ErrCaseNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound, "售后申请不存在"
	case errors.Is(err, afterSalesSvc.ErrCaseWindowExpired):
		return http.StatusConflict, contracts.ErrCodeConflict, "当前不在可申请售后时间窗口"
	default:
		var dup *afterSalesSvc.DuplicateActiveCaseError
		if errors.As(err, &dup) {
			if dup != nil && strings.TrimSpace(dup.CaseNo) != "" {
				return http.StatusConflict, contracts.ErrCodeConflict, "已存在进行中的售后申请：" + dup.CaseNo
			}
			return http.StatusConflict, contracts.ErrCodeConflict, "已存在进行中的售后申请"
		}
		return http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error()
	}
}
