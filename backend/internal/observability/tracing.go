package observability

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName     = "pennywise"
	serviceVersion = "1.0.0"
)

func InitTracer(logger *slog.Logger, traceWriter io.Writer) (func(context.Context) error, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(traceWriter))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("pennywise"),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	logger.Info("tracing initialized")

	return tp.Shutdown, nil
}

func Tracer() trace.Tracer {
	return otel.Tracer(tracerName)
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name)
}
