package logs

import (
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

type ErrorLogger struct {
	log *glog.Logger
}

func NewErrorLogger(loggerPath string) Logger {

	log := glog.New()
	log.SetPath(loggerPath)
	log.SetFile("error.log")
	log.SetPrefix("[error] ")
	log.SetTimeFormat("2006-01-02T15:04:05.000-07:00")
	log.SetHandlers(LoggingJsonHandler)

	return &ErrorLogger{log}

}

func (errorLogger ErrorLogger) Info(ctx context.Context, v ...any) {
	errorLogger.log.Info(ctx, v...)
}

func (errorLogger ErrorLogger) Debug(ctx context.Context, v ...any) {
	errorLogger.log.Debug(ctx, v...)
}

func (errorLogger ErrorLogger) Error(ctx context.Context, v ...any) {
	errorLogger.log.Error(ctx, v...)
}
