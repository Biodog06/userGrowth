package logs

import (
	"os"
	config "usergrowth/configs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type MyLogger struct {
	*zap.Logger
}

//var (
//	loggerInstance *MyLogger
//	obserrvedLogs  *observer.ObservedLogs
//	loggerMutex    sync.Mutex
//)

func InitLogger(loggerPath string) *MyLogger {
	fileWriter := &lumberjack.Logger{
		Filename: loggerPath,
	}
	encoderConfig := zap.NewProductionEncoderConfig() // 默认 json 格式，不用 development
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	ws := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(fileWriter),
	)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		ws,
		zap.InfoLevel,
	)

	return &MyLogger{zap.New(core, zap.AddCaller())}

}

func InitLoggerWithES(loggerPath string, cfg *config.Config) *MyLogger {

	es := NewEsClient(cfg)

	fileWriter := &lumberjack.Logger{
		Filename: loggerPath,
	}
	encoderConfig := zap.NewProductionEncoderConfig() // 默认 json 格式，不用 development
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	ws := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(fileWriter),
		zapcore.AddSync(es),
	)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		ws,
		zap.InfoLevel,
	)

	return &MyLogger{zap.New(core, zap.AddCaller())}
}

func (log *MyLogger) RecordInfoLog(msg string, args ...zap.Field) {
	log.Log(zap.InfoLevel, msg, args...)
}
