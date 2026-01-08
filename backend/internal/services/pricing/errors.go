package pricing

import "errors"

var ErrServiceUnavailable = errors.New("pricing service unavailable")

const (
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeInvalidArgument     = "INVALID_ARGUMENT"
	CodeForbidden           = "FORBIDDEN"
	CodePricebookNotFound   = "PRICEBOOK_NOT_FOUND"
	CodeVersionNotFound     = "VERSION_NOT_FOUND"
	CodeVersionNotEditable  = "VERSION_NOT_EDITABLE"
	CodePublishConflict     = "PUBLISH_CONFLICT"
	CodeInternal            = "INTERNAL_ERROR"
	CodeTenantMissing       = "TENANT_MISSING"
	CodeDuplicatePricebook  = "PRICEBOOK_DUPLICATE"
	CodeInvalidVersionRange = "INVALID_VERSION_RANGE"
)

type Error struct {
	Code string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil && e.Err.Error() != "" {
		return e.Err.Error()
	}
	if e.Code != "" {
		return e.Code
	}
	return "unknown error"
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func E(code string, err error) error {
	if err == nil {
		err = errors.New(code)
	}
	return &Error{Code: code, Err: err}
}
