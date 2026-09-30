package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/gost/errs"
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

	_ = ctx.Error(err)

	e := errs.Normalize(err)
	ctx.AbortWithStatusJSON(e.Status(), Result{
		Code:    e.Code(),
		Message: e.Message(),
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
