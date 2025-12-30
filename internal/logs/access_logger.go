package logs

import (
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

type AccessLogger struct {
	log *glog.Logger
}

func NewAccLogger(loggerPath string) Logger {

	log := glog.New()

	log.SetPath(loggerPath)
	log.SetFile("access.log")
	log.SetPrefix("[access] ")
	log.SetTimeFormat("2006-01-02T15:04:05.000-07:00")
	log.SetHandlers(LoggingJsonHandler)

	return &AccessLogger{log}

}

func (accLog *AccessLogger) Info(ctx context.Context, v ...any) {
	accLog.log.Info(ctx, v...)
}

func (accLog *AccessLogger) Debug(ctx context.Context, v ...any) {
	accLog.log.Debug(ctx, v...)
}

func (accLog *AccessLogger) Error(ctx context.Context, v ...any) {
	accLog.log.Error(ctx, v...)
}
