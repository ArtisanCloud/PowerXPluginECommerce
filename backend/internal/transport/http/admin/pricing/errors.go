package pricing

import (
	"errors"
	"net/http"

	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func respondError(c *gin.Context, status int, code string, err error) {
	if c == nil {
		return
	}
	if status <= 0 {
		status = http.StatusInternalServerError
	}
	msg := "unknown error"
	if err != nil && err.Error() != "" {
		msg = err.Error()
	}
	if code == "" {
		code = "INTERNAL_ERROR"
	}
	c.JSON(status, errorResponse{Error: apiError{Code: code, Message: msg}})
}

func statusFromErr(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func respondServiceError(c *gin.Context, err error) {
	if c == nil {
		return
	}
	status, code := statusAndCodeFromServiceErr(err)
	respondError(c, status, code, err)
}

func statusAndCodeFromServiceErr(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	var se *pricingsvc.Error
	if errors.As(err, &se) {
		switch se.Code {
		case pricingsvc.CodeServiceUnavailable:
			return http.StatusServiceUnavailable, se.Code
		case pricingsvc.CodeForbidden:
			return http.StatusForbidden, se.Code
		case pricingsvc.CodePricebookNotFound, pricingsvc.CodeVersionNotFound:
			return http.StatusNotFound, se.Code
		case pricingsvc.CodeVersionNotEditable, pricingsvc.CodePublishConflict:
			return http.StatusConflict, se.Code
		case pricingsvc.CodeInvalidArgument, pricingsvc.CodeInvalidVersionRange, pricingsvc.CodeDuplicatePricebook, pricingsvc.CodeTenantMissing:
			return http.StatusBadRequest, se.Code
		default:
			return http.StatusInternalServerError, se.Code
		}
	}
	return statusFromErr(err), pricingsvc.CodeInternal
}
