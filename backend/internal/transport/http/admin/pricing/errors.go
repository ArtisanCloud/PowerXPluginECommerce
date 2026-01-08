package pricing

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var ErrServiceUnavailable = errors.New("pricing service unavailable")

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
