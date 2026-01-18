package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func initTracer() (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(getEnv("OTEL_SERVICE_NAME", "tictactoe-backend")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Support both OTLP/gRPC (4317) and OTLP/HTTP (4318) depending on env.
	// If the endpoint includes a URL scheme (http/https), we assume OTLP/HTTP.
	otelEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	otelProtocol := strings.ToLower(strings.TrimSpace(getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "")))

	var traceExporter sdktrace.SpanExporter
	if otelProtocol == "http/protobuf" || strings.HasPrefix(otelEndpoint, "http://") || strings.HasPrefix(otelEndpoint, "https://") {
		// otlptracehttp expects an endpoint without scheme.
		// Allow values like "http://localhost:4318" by stripping the scheme prefix.
		httpEndpoint := strings.TrimPrefix(strings.TrimPrefix(otelEndpoint, "http://"), "https://")
		traceExporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(httpEndpoint),
			otlptracehttp.WithInsecure(),
		)
	} else {
		traceExporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(otelEndpoint),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
