package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	config "usergrowth/configs"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
)

type ErrorManager struct {
	errorLogger logs.Logger
	cfg         *config.MiddlewareConfig
}

func NewErrorManager(loggerPath string, cfg *config.MiddlewareConfig, errorLogger logs.ErrorLogger) *ErrorManager {
	return &ErrorManager{
		errorLogger: errorLogger,
		cfg:         cfg,
	}
}

func (m *ErrorManager) ErrorHandler(r *ghttp.Request) {
	if !m.cfg.Error {
		r.Middleware.Next()
		return
	}
	ctx := r.GetCtx()
	ctx, span := gtrace.NewSpan(ctx, "Middleware.ErrorHandler")
	defer span.End()
	r.SetCtx(ctx)

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
		code := gerror.Code(err)
		switch code {
		case gcode.CodeValidationFailed:
			m.errorLogger.Info(ctx, "validation failed: ", err)
			r.Response.ClearBuffer()
			r.Response.WriteStatus(http.StatusBadRequest)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		case gcode.CodeNotAuthorized:
			m.errorLogger.Info(ctx, "authorization failed: ", err)
			r.Response.ClearBuffer()
			r.Response.WriteStatus(http.StatusUnauthorized)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
				Data:    nil,
			})
		default:
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
