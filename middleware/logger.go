package middleware

import (
	"encoding/json"
	"fmt"
	"time"
	config "usergrowth/configs"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/net/ghttp"
)

type LoggerManager struct {
	accLogger logs.Logger
	cfg       *config.MiddlewareConfig
}

type Content struct {
	AccBody        string        `json:"body"`
	AccMethod      string        `json:"method"`
	AccPath        string        `json:"path"`
	AccIP          string        `json:"ip"`
	AccUA          string        `json:"ua"`
	AllParamString string        `json:"parms"`
	AccResp        int           `json:"resp"`
	AccDura        time.Duration `json:"dura"`
}

func NewLoggerManager(loggerPath string, cfg *config.MiddlewareConfig) *LoggerManager {
	return &LoggerManager{
		accLogger: logs.NewAccLogger(loggerPath),
		cfg:       cfg,
	}
}

func (lm *LoggerManager) AccessHandler(r *ghttp.Request) {
	if !lm.cfg.Access {
		r.Middleware.Next()
		return
	}
	ctx := r.GetCtx()
	accBody := r.GetBodyString()
	accMethod := r.Method
	accPath := r.URL.Path
	accIP := r.GetClientIp()
	accUA := r.UserAgent()
	allParams := r.GetMap()
	var accParamsStr string
	if len(allParams) > 0 {
		if bytes, err := json.Marshal(allParams); err == nil {
			accParamsStr = string(bytes)
		} else {
			accParamsStr = fmt.Sprintf("%v", allParams)
		}
	} else {
		accParamsStr = "{}"
	}

	startTime := time.Now()
	r.Middleware.Next()
	duration := time.Since(startTime)

	accResp := r.Response.Status
	logContent := Content{
		AccBody:        accBody,
		AccMethod:      accMethod,
		AccPath:        accPath,
		AccIP:          accIP,
		AccUA:          accUA,
		AllParamString: accParamsStr,
		AccResp:        accResp,
		AccDura:        duration,
	}
	encodeContent, err := json.Marshal(logContent)
	if err != nil {
		r.SetError(err) // 将错误传递给上下文，以便 ErrorHandler 捕获
		return
	}

	lm.accLogger.Debug(ctx, encodeContent)
	fmt.Println(string(encodeContent))
}
