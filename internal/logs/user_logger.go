package logs

import (
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

type UserLogger struct {
	log *glog.Logger
}

func NewUserLogger(loggerPath string) Logger {

	log := glog.New()

	log.SetPath(loggerPath)
	log.SetFile("user.log")
	log.SetPrefix("[user] ")
	log.SetTimeFormat("2006-01-02T15:04:05.000-07:00")
	log.SetHandlers(LoggingJsonHandler)

	return &UserLogger{log}

}

func (userLog *UserLogger) Info(ctx context.Context, v ...any) {
	userLog.log.Info(ctx, v...)
}

func (userLog *UserLogger) Debug(ctx context.Context, v ...any) {
	userLog.log.Debug(ctx, v...)
}

func (userLog *UserLogger) Error(ctx context.Context, v ...any) {
	userLog.log.Error(ctx, v...)
}

func (userLog *UserLogger) Fatal(ctx context.Context, v ...any) {
	userLog.log.Fatal(ctx, v...)
}
