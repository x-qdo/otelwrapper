package otelwrapper

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// InitTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the exporters (OTLP by default) that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func InitTracerProvider(serviceName, namespace string, exporters ...tracesdk.SpanExporter) (
	*tracesdk.TracerProvider, error,
) {
	opts := make([]tracesdk.TracerProviderOption, 0)

	if len(exporters) == 0 {
		client := otlptracegrpc.NewClient()
		exp, err := otlptrace.New(context.Background(), client)
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, exp)
	}

	// Always be sure to batch in production.
	for _, exp := range exporters {
		opts = append(opts, tracesdk.WithBatcher(exp))
	}

	// Record information about this application in a Resource.
	opts = append(
		opts, tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				semconv.ServiceNamespaceKey.String(namespace),
			),
		),
	)
	tp := tracesdk.NewTracerProvider(opts...)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))

	return tp, nil
}

func ShutdownWaiting(tp *tracesdk.TracerProvider, ctx context.Context, wg *sync.WaitGroup) {
	childCtx := context.WithValue(ctx, "tracer provider shutdown", nil)
	wg.Add(1)
	<-childCtx.Done()
	fmt.Println("Shutting down OTel tracer provider...")

	if err := tp.Shutdown(context.Background()); err != nil {
		fmt.Printf("OTel tracer provider forced to shutdown: %v\n", err)
	} else {
		fmt.Println("OTel tracer provider shutdown.")
	}

	wg.Done()
}
