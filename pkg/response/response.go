package response

import (
	"net/http"

	"isac-cran-system/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    int(errors.CodeSuccess),
		Message: errors.CodeSuccess.Message(),
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    int(errors.CodeSuccess),
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	var appErr *errors.AppError
	if ae, ok := err.(*errors.AppError); ok {
		appErr = ae
	} else {
		appErr = errors.InternalError("internal server error", err)
	}

	c.JSON(appErr.HTTPStatus(), Response{
		Code:    appErr.Code.Int(),
		Message: appErr.Message,
	})
}

func ErrorWithCode(c *gin.Context, code errors.Code, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code.Int(),
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    int(errors.CodeBadRequest),
		Message: message,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    int(errors.CodeUnauthorized),
		Message: message,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    int(errors.CodeForbidden),
		Message: message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    int(errors.CodeNotFound),
		Message: message,
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    int(errors.CodeInternalError),
		Message: message,
	})
}

func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
