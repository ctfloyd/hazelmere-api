package monitor

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds application-level metrics instruments.
// Add counters, histograms, and gauges here as needed.
type Metrics struct {
	meter metric.Meter

	// Example metrics (stubbed for now):
	// RequestCount   metric.Int64Counter
	// RequestLatency metric.Float64Histogram
}

// NewMetrics creates and registers application metrics.
func NewMetrics() *Metrics {
	meter := otel.Meter("hazelmere")

	// Initialize metrics here as needed:
	// requestCount, _ := meter.Int64Counter("http.request.count",
	// 	metric.WithDescription("Total number of HTTP requests"),
	// )
	// requestLatency, _ := meter.Float64Histogram("http.request.latency",
	// 	metric.WithDescription("HTTP request latency in milliseconds"),
	// )

	return &Metrics{
		meter: meter,
		// RequestCount:   requestCount,
		// RequestLatency: requestLatency,
	}
}

// Meter returns the underlying OpenTelemetry meter for creating custom metrics.
func (m *Metrics) Meter() metric.Meter {
	return m.meter
}
