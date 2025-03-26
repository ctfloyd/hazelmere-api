package initialize

import (
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"time"
)

func InitRouter(log hz_logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(1 * time.Second))
	router.Use(middleware.RequestID)
	router.Use(hz_logger.NewMiddleware(log).Serve)
	return router
}
