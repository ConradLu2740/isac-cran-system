package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, message=%s, detail=%s, error=%v",
			e.Code, e.Message, e.Detail, e.Err)
	}
	return fmt.Sprintf("code=%d, message=%s, detail=%s",
		e.Code, e.Message, e.Detail)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithDetail(code Code, message string, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

func Wrap(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func WrapWithDetail(code Code, message string, detail string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
		Err:     err,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:    CodeBadRequest,
		Message: message,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:    CodeForbidden,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: message,
	}
}

func InternalError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code >= 20001 && e.Code < 30000:
		return http.StatusBadGateway
	case e.Code >= 40001 && e.Code < 50000:
		return http.StatusServiceUnavailable
	case e.Code == CodeUnauthorized:
		return http.StatusUnauthorized
	case e.Code == CodeForbidden:
		return http.StatusForbidden
	case e.Code == CodeNotFound:
		return http.StatusNotFound
	case e.Code >= 10001 && e.Code < 20000:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func IsCode(err error, code Code) bool {
	if err == nil {
		return false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

func GetCode(err error) Code {
	if err == nil {
		return CodeSuccess
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return CodeInternalError
}
