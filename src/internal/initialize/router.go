package initialize

import (
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func InitRouter(log hz_logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.AllowCors)
	router.Use(chiWare.Recoverer)
	router.Use(chiWare.RequestID)
	router.Use(otelhttp.NewMiddleware("hazelmere"))
	router.Use(hz_logger.NewMiddleware(log).Serve)
	return router
}
