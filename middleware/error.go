package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ErrorManager struct {
	errorLogger logs.Logger
}

func NewErrorManager(loggerPath string) *ErrorManager {
	return &ErrorManager{logs.NewErrorLogger(loggerPath)}
}

func (m *ErrorManager) ErrorHandler(r *ghttp.Request) {
	ctx := r.GetCtx()
	defer func() {
		if exception := recover(); exception != nil {
			errorMsg := fmt.Sprintf("SERVER PANIC: %v\n%s", exception, debug.Stack())
			m.errorLogger.Error(ctx, errorMsg)
			r.Response.ClearBuffer()

			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.WriteJson(
				ghttp.DefaultHandlerResponse{
					Code:    http.StatusInternalServerError,
					Message: "服务器繁忙，请稍后再试",
					Data:    nil,
				})
		}
	}()
	r.Middleware.Next()
	err := r.GetError()

	if err != nil {
		if gerror.Code(err) == gcode.CodeValidationFailed {
			m.errorLogger.Info(ctx, "validation failed: ", err)
			r.Response.ClearBuffer()
			r.Response.WriteStatus(http.StatusBadRequest)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		} else {
			m.errorLogger.Error(ctx, "internal error: ", err)
			// 如果 Controller 还没有写入响应，则返回默认错误
			if r.Response.BufferLength() == 0 {
				r.Response.ClearBuffer()
				r.Response.WriteStatus(http.StatusInternalServerError)
				r.Response.WriteJson(
					ghttp.DefaultHandlerResponse{
						Code:    http.StatusInternalServerError,
						Message: "服务器繁忙，请稍后再试",
						Data:    nil,
					})
			}
		}
	}
}
