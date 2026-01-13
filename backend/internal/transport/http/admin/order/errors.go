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
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
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
		errors.Is(err, ordersvc.ErrItemsRequired),
		errors.Is(err, ordersvc.ErrInvalidQty),
		errors.Is(err, ordersvc.ErrDuplicateSKU):
		return http.StatusBadRequest, contracts.ErrCodeInvalidRequest
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
