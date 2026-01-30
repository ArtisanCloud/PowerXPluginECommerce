package payments

import (
	"errors"
	"io"
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
	case errors.Is(err, paymentssvc.ErrOrderCustomerMismatch):
		return http.StatusForbidden, contracts.ErrCodeInvalidRequest, "订单归属不匹配"
	case errors.Is(err, paymentssvc.ErrCustomerRequired):
		return http.StatusUnauthorized, contracts.ErrCodeInvalidRequest, "缺少客户信息"
	case errors.Is(err, paymentssvc.ErrCustomerIdentityNotFound):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "缺少微信身份信息"
	case errors.Is(err, paymentssvc.ErrProviderUnavailable):
		return http.StatusUnprocessableEntity, contracts.ErrCodeInvalidRequest, "支付渠道暂不可用"
	case errors.Is(err, paymentssvc.ErrProviderSelectorRequired):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "缺少支付渠道定位信息（providerId）"
	case errors.Is(err, paymentssvc.ErrTransactionNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound, "支付单不存在"
	case errors.Is(err, paymentssvc.ErrOpenIDRequired):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "缺少 openid"
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

func sendProviderCallbackResponse(c *gin.Context, resp *http.Response) error {
	if c == nil || resp == nil {
		return nil
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode > 0 {
			c.Writer.WriteHeader(resp.StatusCode)
		}
		if len(body) > 0 {
			_, err = c.Writer.Write(body)
			return err
		}
		return nil
	}
	if resp.StatusCode > 0 {
		c.Writer.WriteHeader(resp.StatusCode)
	}
	return nil
}
