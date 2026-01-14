package order

import (
	"errors"
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	ordersvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/order"
	"github.com/gin-gonic/gin"
)

func respondAdminOrderError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	status, code := httpStatusForOrderError(err)
	msg := messageForOrderError(err)
	contracts.ResponseError(c, status, code, msg)
}

func httpStatusForOrderError(err error) (status int, code string) {
	if err == nil {
		return http.StatusOK, ""
	}
	switch {
	case errors.Is(err, ordersvc.ErrOrderServiceUnavailable):
		return http.StatusServiceUnavailable, contracts.ErrCodeInternalError
	case errors.Is(err, ordersvc.ErrAdminRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized
	case errors.Is(err, ordersvc.ErrOrderNotFound),
		errors.Is(err, ordersvc.ErrCustomerNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired),
		errors.Is(err, ordersvc.ErrCustomerRequired),
		errors.Is(err, ordersvc.ErrChannelRequired),
		errors.Is(err, ordersvc.ErrShippingAddressRequired),
		errors.Is(err, ordersvc.ErrInvalidShippingAddress),
		errors.Is(err, ordersvc.ErrItemsRequired),
		errors.Is(err, ordersvc.ErrInvalidQty),
		errors.Is(err, ordersvc.ErrDuplicateSKU):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
	case errors.Is(err, ordersvc.ErrShippingAddressNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	case errors.Is(err, ordersvc.ErrIdempotencyConflict),
		errors.Is(err, ordersvc.ErrIdempotencyInProgress),
		errors.Is(err, ordersvc.ErrOrderNotCancellable),
		errors.Is(err, ordersvc.ErrOutOfStock),
		errors.Is(err, ordersvc.ErrSellabilityFailed):
		return http.StatusConflict, contracts.ErrCodeConflict
	default:
		return http.StatusInternalServerError, contracts.ErrCodeInternalError
	}
}

func messageForOrderError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ordersvc.ErrOrderServiceUnavailable):
		return "订单服务不可用"
	case errors.Is(err, ordersvc.ErrAdminRequired):
		return "未授权"
	case errors.Is(err, ordersvc.ErrOrderNotFound):
		return "订单不存在"
	case errors.Is(err, ordersvc.ErrCustomerNotFound):
		return "客户不存在"
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired):
		return "缺少幂等键（Idempotency-Key）"
	case errors.Is(err, ordersvc.ErrIdempotencyConflict):
		return "幂等键冲突：请求参数与历史不一致"
	case errors.Is(err, ordersvc.ErrIdempotencyInProgress):
		return "请求处理中，请稍后重试"
	case errors.Is(err, ordersvc.ErrCustomerRequired):
		return "customerId 必填"
	case errors.Is(err, ordersvc.ErrChannelRequired):
		return "channel 必填"
	case errors.Is(err, ordersvc.ErrShippingAddressRequired):
		return "收货地址必填"
	case errors.Is(err, ordersvc.ErrInvalidShippingAddress):
		return "收货地址不完整"
	case errors.Is(err, ordersvc.ErrShippingAddressNotFound):
		return "收货地址不存在"
	case errors.Is(err, ordersvc.ErrItemsRequired):
		return "items 不能为空"
	case errors.Is(err, ordersvc.ErrInvalidQty):
		return "购买数量必须大于 0"
	case errors.Is(err, ordersvc.ErrDuplicateSKU):
		return "同一订单中 SKU 不可重复"
	case errors.Is(err, ordersvc.ErrSellabilityFailed):
		return "商品不可售"
	case errors.Is(err, ordersvc.ErrOutOfStock):
		return "库存不足"
	case errors.Is(err, ordersvc.ErrOrderNotCancellable):
		return "当前订单状态不允许取消"
	default:
		return err.Error()
	}
}
