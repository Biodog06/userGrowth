package observability

import (
	"context"
	"time"
	"usergrowth/internal/logs"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(serviceName, endpoint, path string, errorLogger logs.Logger) func() {

	exporter, err1 := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithInsecure(), // 本地测试禁用 TLS，生产环境需配置
	)
	ctx := context.Background()
	if err1 != nil {
		errorLogger.Fatal(ctx, "创建 OTLP exporter 失败: %v", err1)
	}
	sampler := sdktrace.AlwaysSample()

	res, err2 := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	if err2 != nil {
		errorLogger.Fatal(ctx, "创建 Resource 失败: %v", err2)
		return nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			errorLogger.Fatal(ctx, "TracerProvider 关闭失败: %v", err)
		}
	}
}
