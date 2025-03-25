package metrics

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"net/http"
	"time"
)

var meter = otel.Meter("middleware")

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		metricPrefix := fmt.Sprintf("hazelmere-api.prod.us-west-2.%s.%s", r.Method, r.URL.Path)

		t1 := time.Now()
		defer func() {
			timer, err := meter.Int64Gauge(fmt.Sprintf("%s.Time.Millis", metricPrefix))
			if err == nil {
				timer.Record(context.Background(), time.Since(t1).Milliseconds())
			}

			statusString := fmt.Sprintf("%dXX", ww.Status()/100)
			status, err := meter.Int64Counter(fmt.Sprintf("%s.Status.%s", metricPrefix, statusString))
			if err == nil {
				status.Add(context.Background(), 1)
			}
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
