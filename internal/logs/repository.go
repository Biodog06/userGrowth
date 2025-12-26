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
	es *MyAsyncEs
}

func (log *MyLogger) RecordInfoLog(msg string, args ...zap.Field) {
	log.Log(zap.InfoLevel, msg, args...)
}

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

	return &MyLogger{zap.New(core, zap.AddCaller()), nil}

}

func InitLoggerWithES(loggerPath string, cfg *config.Config, es *MyAsyncEs) *MyLogger {

	fileWriter := &lumberjack.Logger{
		Filename: loggerPath,
	}

	encoderConfig := zap.NewProductionEncoderConfig() // 默认 json 格式，不用 development
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	ws := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(fileWriter),
		es,
	)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		ws,
		zap.InfoLevel,
	)

	return &MyLogger{zap.New(core, zap.AddCaller()), es}
}

// SyncVersion of ES

//type MyEs struct {
//	client *elastic.Client
//	index  string
//}

//func NewEsClient(cfg *config.Config) *MyEs {
//	esCfg := elastic.Config{
//		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.ES.Host, strconv.Itoa(cfg.ES.Port))},
//	}
//	es, err := elastic.NewClient(esCfg)
//	if err != nil {
//		panic(err)
//	}
//	return &MyEs{client: es, index: "zap-logs"}
//}

//func (es *MyEs) Write(p []byte) (n int, err error) {
//	res, err := es.client.Index(
//		es.index,
//		bytes.NewReader(p),
//		es.client.Index.WithContext(context.Background()),
//	)
//	if err != nil {
//		fmt.Println(err)
//		return 0, err
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			fmt.Println("not close:", err)
//		}
//	}(res.Body)
//	return len(p), nil
//}
