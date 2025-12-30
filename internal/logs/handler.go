package logs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
)

type JsonOutputsForLogger struct {
	Timestamp string `json:"@timestamp"`
	TraceId   string `json:"trace_id"`
	Level     string `json:"level"`
	Content   string `json:"content"`
}

var LoggingJsonHandler glog.Handler = func(ctx context.Context, in *glog.HandlerInput) {
	jsonForLogger := JsonOutputsForLogger{
		Timestamp: in.TimeFormat,
		Level:     gstr.Trim(in.LevelFormat, "[]"),
		Content:   gstr.Trim(in.ValuesContent()),
	}
	jsonBytes, err := json.Marshal(jsonForLogger)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		return
	}
	in.Buffer.Write(jsonBytes)
	in.Buffer.WriteString("\n")
	in.Next(ctx)
}
