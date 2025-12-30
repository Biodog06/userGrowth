package middleware

import (
	"encoding/json"
	"fmt"
	"time"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/net/ghttp"
)

type LoggerManager struct {
	accLogger logs.Logger
}

func NewLoggerManager(loggerPath string) *LoggerManager {
	return &LoggerManager{logs.NewAccLogger(loggerPath)}
}

func (lm *LoggerManager) AccessHandler(r *ghttp.Request) {
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
	logContent := fmt.Sprintf(
		"method:%s path:%s status:%d ip:%s cost:%v params:%s body:%s ua:%s",
		accMethod,
		accPath,
		accResp,
		accIP,
		duration,     // 耗时
		accParamsStr, // 这里是 {"username":"..."} 这样的字符串
		accBody,      // 请求体
		accUA,
	)

	// 7. 写入日志
	lm.accLogger.Debug(ctx, logContent)

}
