package handler

import (
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type Health struct {
	Status string `json:"status"`
}

type HealthHandler struct {
	logger hz_logger.Logger
}

func NewHealthHandler(logger hz_logger.Logger) *HealthHandler {
	return &HealthHandler{logger}
}

func (hh *HealthHandler) RegisterRoutes(mux *chi.Mux, version ApiVersion, authorizer *middleware.Authorizer) {
	if version == ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(100 * time.Millisecond))
			r.Get("/health", hh.HealthCheck)
		})
	}
}

func (hh *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	hz_handler.Ok(w, Health{Status: "ok"})
}
