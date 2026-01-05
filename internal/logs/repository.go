package logs

import (
	"context"
)

type Logger interface {
	Info(ctx context.Context, v ...any)
	Debug(ctx context.Context, v ...any)
	Error(ctx context.Context, v ...any)
	Fatal(ctx context.Context, v ...any)
}
