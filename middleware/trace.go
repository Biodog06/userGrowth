package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func Trace(r *ghttp.Request) {
	ctx := r.Context()
	tr := otel.Tracer("gf-http")

	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	spanName := r.URL.Path
	if r.Router != nil {
		spanName = r.Router.Uri
	}
	ctx, span := tr.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	r.SetCtx(ctx)

	span.SetAttributes(
		semconv.HTTPMethodKey.String(r.Method),
		semconv.HTTPTargetKey.String(r.URL.String()),
		semconv.NetHostNameKey.String(r.Host),
		semconv.HTTPClientIPKey.String(r.GetClientIp()),
	)

	r.Middleware.Next()

	status := r.Response.Status
	span.SetAttributes(
		semconv.HTTPStatusCodeKey.Int(status),
	)

	if status >= 500 {
		span.RecordError(r.GetError())
	}
}
