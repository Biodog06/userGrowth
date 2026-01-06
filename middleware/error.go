package middleware

import (
	"net/http"
	config "usergrowth/configs"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
)

type ErrorManager struct {
	errorLogger logs.Logger
	cfg         *config.MiddlewareConfig
}

func NewErrorManager(loggerPath string, cfg *config.MiddlewareConfig, errorLogger logs.Logger) *ErrorManager {
	return &ErrorManager{
		errorLogger: errorLogger,
		cfg:         cfg,
	}
}

func (m *ErrorManager) ErrorHandler(r *ghttp.Request) {
	if m.cfg.Error == nil || !*m.cfg.Error {
		r.Middleware.Next()
		return
	}
	ctx := r.GetCtx()
	ctx, span := gtrace.NewSpan(ctx, "Middleware.ErrorHandler")
	defer span.End()
	r.SetCtx(ctx)

	r.Middleware.Next()
	err := r.GetError()

	if err != nil {
		code := gerror.Code(err)
		switch code {
		case gcode.CodeValidationFailed:
			m.errorLogger.Info(ctx, "validation failed: ", err)
			r.Response.ClearBuffer()
			r.Response.WriteJson(g.Map{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
				"data":    nil,
			})
		case gcode.CodeNotAuthorized:
			m.errorLogger.Info(ctx, "authorization failed: ", err)
			r.Response.ClearBuffer()
			r.Response.WriteJson(g.Map{
				"code":    http.StatusUnauthorized,
				"message": err.Error(),
				"data":    nil,
			})
		default:
			isPanic := false
			if stack := gerror.Stack(err); stack != "" {
				isPanic = true
			}
			if isPanic {
				stack := gerror.Stack(err)
				m.errorLogger.Error(ctx,
					"PANIC recovered by framework",
					"error", err.Error(),
					g.Map{"stack": stack},
				)
				r.Response.ClearBuffer()
				r.Response.WriteJson(g.Map{
					"code":    http.StatusInternalServerError,
					"message": "服务器繁忙，请稍后再试",
					// "debug_stack": stack, // 仅 debug 环境开启
				})
			} else {
				m.errorLogger.Error(ctx, "internal error: ", err.Error())
				r.Response.ClearBuffer()
				r.Response.WriteJson(g.Map{
					"code":    http.StatusInternalServerError,
					"message": "服务器繁忙，请稍后再试",
				})
			}
		}
	}
}
