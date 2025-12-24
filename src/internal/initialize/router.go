package initialize

import (
	"net/http"
	"regexp"

	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

func InitRouter(log hz_logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.AllowCors)
	router.Use(chiWare.Recoverer)
	router.Use(chiWare.RequestID)
	router.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewMiddleware("hazelmere")(next)
	})
	router.Use(chiRouteLabeler) // runs after otelhttp, adds route label
	router.Use(hz_logger.NewMiddleware(log).Serve)
	return router
}

// routeRegexPattern matches chi route params with regex patterns like {userId:[0-9a-fA-F-]+}
var routeRegexPattern = regexp.MustCompile(`\{(\w+):[^}]+\}`)

// chiRouteLabeler adds the chi route pattern to otelhttp metrics/spans
func chiRouteLabeler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		// After handler runs, chi has resolved the route
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern != "" {
			// Simplify route pattern: {userId:[0-9a-fA-F-]+} -> {userId}
			cleanRoute := routeRegexPattern.ReplaceAllString(routePattern, "{$1}")
			labeler, _ := otelhttp.LabelerFromContext(r.Context())
			labeler.Add(attribute.String("http.route", cleanRoute))
		}
	})
}
