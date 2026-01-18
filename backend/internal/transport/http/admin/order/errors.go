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
	case errors.Is(err, ordersvc.ErrBenefitReviewServiceUnavailable):
		return http.StatusServiceUnavailable, contracts.ErrCodeInternalError
	case errors.Is(err, ordersvc.ErrAdminRequired):
		return http.StatusUnauthorized, contracts.ErrCodeUnauthorized
	case errors.Is(err, ordersvc.ErrOrderNotFound),
		errors.Is(err, ordersvc.ErrCustomerNotFound),
		errors.Is(err, ordersvc.ErrBenefitReviewNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	case errors.Is(err, ordersvc.ErrIdempotencyKeyRequired),
		errors.Is(err, ordersvc.ErrCustomerRequired),
		errors.Is(err, ordersvc.ErrChannelRequired),
		errors.Is(err, ordersvc.ErrShippingAddressRequired),
		errors.Is(err, ordersvc.ErrInvalidShippingAddress),
		errors.Is(err, ordersvc.ErrItemsRequired),
		errors.Is(err, ordersvc.ErrInvalidQty),
		errors.Is(err, ordersvc.ErrDuplicateSKU),
		errors.Is(err, ordersvc.ErrBenefitReviewCodeRequired),
		errors.Is(err, ordersvc.ErrBenefitReviewTypeInvalid),
		errors.Is(err, ordersvc.ErrBenefitReviewValueInvalid),
		errors.Is(err, ordersvc.ErrBenefitReviewReasonRequired):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
	case errors.Is(err, ordersvc.ErrShippingAddressNotFound):
		return http.StatusNotFound, contracts.ErrCodeNotFound
	case errors.Is(err, ordersvc.ErrIdempotencyConflict),
		errors.Is(err, ordersvc.ErrIdempotencyInProgress),
		errors.Is(err, ordersvc.ErrOrderNotCancellable),
		errors.Is(err, ordersvc.ErrOrderNotEditable),
		errors.Is(err, ordersvc.ErrBenefitReviewInvalidStatus),
		errors.Is(err, ordersvc.ErrBenefitReviewSameOperator),
		errors.Is(err, ordersvc.ErrBenefitReviewOrderNotEditable),
		errors.Is(err, ordersvc.ErrBenefitReviewUsed),
		errors.Is(err, ordersvc.ErrBenefitReviewTotalExceeded),
		errors.Is(err, ordersvc.ErrBenefitReviewStackingNotAllowed),
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
	case errors.Is(err, ordersvc.ErrOrderNotEditable):
		return "当前订单状态不允许编辑"
	case errors.Is(err, ordersvc.ErrBenefitReviewServiceUnavailable):
		return "优惠权益审核服务不可用"
	case errors.Is(err, ordersvc.ErrBenefitReviewNotFound):
		return "优惠权益审核记录不存在"
	case errors.Is(err, ordersvc.ErrBenefitReviewInvalidStatus):
		return "优惠权益审核状态不可用"
	case errors.Is(err, ordersvc.ErrBenefitReviewSameOperator):
		return "审核人不能与提交人相同"
	case errors.Is(err, ordersvc.ErrBenefitReviewReasonRequired):
		return "拒绝原因必填"
	case errors.Is(err, ordersvc.ErrBenefitReviewOrderNotEditable):
		return "当前订单状态不允许使用优惠权益"
	case errors.Is(err, ordersvc.ErrBenefitReviewCodeRequired):
		return "权益码必填"
	case errors.Is(err, ordersvc.ErrBenefitReviewTypeInvalid):
		return "权益类型不合法"
	case errors.Is(err, ordersvc.ErrBenefitReviewValueInvalid):
		return "优惠值不合法"
	case errors.Is(err, ordersvc.ErrBenefitReviewUsed):
		return "该权益码已被使用"
	case errors.Is(err, ordersvc.ErrBenefitReviewTotalExceeded):
		return "抵扣金额不能超过订单金额"
	case errors.Is(err, ordersvc.ErrBenefitReviewStackingNotAllowed):
		return "当前订单不允许叠加优惠"
	default:
		return err.Error()
	}
}
