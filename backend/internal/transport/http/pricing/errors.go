package pricing

import (
	"errors"
	"net/http"

	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/gin-gonic/gin"
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
		code = pricingsvc.CodeInternal
	}
	c.JSON(status, errorResponse{Error: apiError{Code: code, Message: msg}})
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
		case pricingsvc.CodeInvalidArgument, pricingsvc.CodeInvalidVersionRange, pricingsvc.CodeTenantMissing:
			return http.StatusBadRequest, se.Code
		case pricingsvc.CodeServiceUnavailable:
			return http.StatusServiceUnavailable, se.Code
		default:
			return http.StatusInternalServerError, se.Code
		}
	}
	return http.StatusInternalServerError, pricingsvc.CodeInternal
}
