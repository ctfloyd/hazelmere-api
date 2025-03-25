package initialize

import (
	"api/src/internal/common/logger"
	"api/src/internal/common/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"time"
)

func InitRouter(log logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(1 * time.Second))
	router.Use(middleware.RequestID)
	router.Use(metrics.Middleware)
	router.Use(logger.NewMiddleware(log).Serve)
	return router
}
