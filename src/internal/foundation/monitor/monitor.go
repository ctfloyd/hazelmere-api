package monitor

import (
	"context"

	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Monitor provides unified access to observability primitives: tracing, logging, and metrics.
type Monitor struct {
	tracer  trace.Tracer
	logger  hz_logger.Logger
	metrics *Metrics
}

// New creates a new Monitor with the given logger.
func New(logger hz_logger.Logger) *Monitor {
	return &Monitor{
		tracer:  otel.Tracer("hazelmere"),
		logger:  logger,
		metrics: NewMetrics(),
	}
}

// Tracer returns the OpenTelemetry tracer.
func (m *Monitor) Tracer() trace.Tracer {
	return m.tracer
}

// Logger returns the application logger.
func (m *Monitor) Logger() hz_logger.Logger {
	return m.logger
}

// Metrics returns the application metrics.
func (m *Monitor) Metrics() *Metrics {
	return m.metrics
}

// StartSpan starts a new span with the given name and returns the updated context and span.
func (m *Monitor) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return m.tracer.Start(ctx, name)
}
