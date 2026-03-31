package contracts

import (
	"errors"
	"net/http"
	"strings"

	adminAfterSales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	"gorm.io/gorm"
)

type ErrorMapping struct {
	Status  int
	Code    string
	Message string
}

// MapAfterSalesError maps domain errors to unified HTTP/code/message triples.
func MapAfterSalesError(err error) ErrorMapping {
	if err == nil {
		return ErrorMapping{Status: http.StatusInternalServerError, Code: ErrCodeInternalError, Message: "unknown error"}
	}
	switch {
	case errors.Is(err, adminAfterSales.ErrAdminServiceUnavailable),
		errors.Is(err, adminAfterSales.ErrPaymentGuardUnavailable),
		errors.Is(err, adminAfterSales.ErrOrderSyncUnavailable),
		errors.Is(err, adminAfterSales.ErrReverseLinkUnavailable):
		return ErrorMapping{Status: http.StatusServiceUnavailable, Code: ErrCodeAfterSalesServiceUnavailable, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrAdminCaseNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return ErrorMapping{Status: http.StatusNotFound, Code: ErrCodeAfterSalesCaseNotFound, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrDecisionReasonRequired),
		errors.Is(err, adminAfterSales.ErrInvalidAction),
		errors.Is(err, adminAfterSales.ErrReverseWaybillRequired),
		errors.Is(err, adminAfterSales.ErrReverseLinkNotAllowed):
		code := ErrCodeAfterSalesInvalidAction
		if errors.Is(err, adminAfterSales.ErrDecisionReasonRequired) {
			code = ErrCodeAfterSalesReasonRequired
		}
		if errors.Is(err, adminAfterSales.ErrReverseLinkNotAllowed) {
			code = ErrCodeAfterSalesReverseNotAllowed
		}
		return ErrorMapping{Status: http.StatusBadRequest, Code: code, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrCoreFieldsFrozen):
		return ErrorMapping{Status: http.StatusConflict, Code: ErrCodeAfterSalesCoreFieldsFrozen, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrRefundAlreadyApplied),
		errors.Is(err, adminAfterSales.ErrRefundAlreadyInPayment):
		return ErrorMapping{Status: http.StatusConflict, Code: ErrCodeAfterSalesRefundDuplicated, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrReverseWaybillNotFound):
		return ErrorMapping{Status: http.StatusNotFound, Code: ErrCodeAfterSalesReverseNotFound, Message: err.Error()}
	case errors.Is(err, adminAfterSales.ErrReverseWaybillOrderMismatch):
		return ErrorMapping{Status: http.StatusConflict, Code: ErrCodeAfterSalesReverseMismatch, Message: err.Error()}
	default:
		msg := strings.TrimSpace(err.Error())
		if strings.Contains(msg, "invalid status transition") {
			return ErrorMapping{Status: http.StatusConflict, Code: ErrCodeAfterSalesInvalidState, Message: msg}
		}
		return ErrorMapping{Status: http.StatusInternalServerError, Code: ErrCodeInternalError, Message: msg}
	}
}
