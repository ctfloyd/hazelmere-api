package initialize

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// OtelConfig holds OpenTelemetry configuration
type OtelConfig struct {
	Enabled          bool
	ServiceName      string
	ServiceNamespace string
	Endpoint         string
	AuthHeader       string
}

// InitOtel initializes OpenTelemetry tracing and metrics.
// When disabled, it uses noop providers. When enabled, it exports to the configured endpoint.
// Returns a shutdown function that should be deferred.
func InitOtel(ctx context.Context, cfg OtelConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		// Disabled: no-op, just return empty shutdown
		return func(context.Context) error { return nil }, nil
	}

	// Set up resource with service info
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.ServiceNamespace),
		),
	)
	if err != nil {
		return nil, err
	}

	// Build exporter options
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(cfg.Endpoint),
	}
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(cfg.Endpoint),
	}

	if cfg.AuthHeader != "" {
		headers := map[string]string{"Authorization": cfg.AuthHeader}
		traceOpts = append(traceOpts, otlptracehttp.WithHeaders(headers))
		metricOpts = append(metricOpts, otlpmetrichttp.WithHeaders(headers))
	}

	// Trace exporter
	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Metric exporter
	metricExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// Set up propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		var errs []error
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	}

	return shutdown, nil
}
