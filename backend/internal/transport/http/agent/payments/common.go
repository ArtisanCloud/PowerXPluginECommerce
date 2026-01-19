package payments

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	paymentssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	"github.com/gin-gonic/gin"
)

func respondPaymentError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	logger.WithError(err).WithFields(logger.Fields{
		"request_id": requestIDFromRequest(c),
		"path":       c.FullPath(),
		"method":     c.Request.Method,
	}).Error("agent payments request failed")
	status, code, msg := httpStatusForPaymentError(err)
	contracts.ResponseError(c, status, code, msg)
}

func httpStatusForPaymentError(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "", ""
	}
	switch {
	case errors.Is(err, paymentssvc.ErrPaymentServiceUnavailable):
		return http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "支付服务不可用"
	case errors.Is(err, paymentssvc.ErrInvalidArgument):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "请求参数不完整"
	case errors.Is(err, paymentssvc.ErrOrderNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound, "订单不存在"
	case errors.Is(err, paymentssvc.ErrOrderNotPayable):
		return http.StatusConflict, contracts.ErrCodeConflict, "订单不可支付"
	case errors.Is(err, paymentssvc.ErrProviderUnavailable):
		return http.StatusUnprocessableEntity, contracts.ErrCodeInvalidRequest, "支付渠道暂不可用"
	case errors.Is(err, paymentssvc.ErrTransactionNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound, "支付单不存在"
	default:
		return http.StatusInternalServerError, contracts.ErrCodeInternalError, strings.TrimSpace(err.Error())
	}
}

func requestIDFromRequest(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v := strings.TrimSpace(c.GetHeader("X-Request-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.GetHeader("Request-ID")); v != "" {
		return v
	}
	return ""
}
