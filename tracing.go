package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var tracer trace.Tracer

func initTracer() (func(context.Context) error, error) {
	// Create stdout exporter
	// stdoutExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	// }

	// Create OTLP gRPC exporter
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	otlpExp, err := otlptracegrpc.New(ctx,
		// otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("nfl-scores"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		// sdktrace.WithBatcher(stdoutExp),
		sdktrace.WithBatcher(otlpExp),
		sdktrace.WithResource(res),
	)

	// otel.SetTracerProvider(tp)
	otel.SetTracerProvider(noop.NewTracerProvider()) // TODO: Remove this
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Set global tracer
	tracer = tp.Tracer("nfl-scores")

	return tp.Shutdown, nil
}
