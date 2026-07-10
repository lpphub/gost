package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Result{
		Code:    0,
		Message: "ok",
	})
}

func OKWithData(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Result{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Fail(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	if bizErr, ok := errors.AsType[*Error](err); ok {
		failWithBizErr(ctx, bizErr)
		return
	}

	fail(ctx, http.StatusOK, -1, err.Error())
}

func failWithBizErr(ctx *gin.Context, err *Error) {
	status := err.HTTPStatus
	if status == 0 {
		status = http.StatusOK
	}

	fail(ctx, status, err.Code, err.Message)
}

func fail(ctx *gin.Context, httpStatus int, code int, msg string) {
	ctx.AbortWithStatusJSON(httpStatus, Result{
		Code:    code,
		Message: msg,
	})
}

func FailWithData(ctx *gin.Context, err error, data any) {
	if bizErr, ok := errors.AsType[*Error](err); ok {
		failWithBizErrData(ctx, bizErr, data)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, Result{
		Code:    -1,
		Message: err.Error(),
		Data:    data,
	})
}

func failWithBizErrData(ctx *gin.Context, err *Error, data any) {
	status := err.HTTPStatus
	if status == 0 {
		status = http.StatusOK
	}
	ctx.AbortWithStatusJSON(status, Result{
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	})
}

func Respond(ctx *gin.Context, err error, data ...any) {
	if err != nil {
		Fail(ctx, err)
		return
	}

	if len(data) == 0 {
		OK(ctx)
		return
	}

	OKWithData(ctx, data[0])
}
