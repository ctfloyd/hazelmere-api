package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/health"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type HealthHandler struct {
	logger        hz_logger.Logger
	healthService *health.Service
}

func NewHealthHandler(logger hz_logger.Logger, healthService *health.Service) *HealthHandler {
	return &HealthHandler{
		logger:        logger,
		healthService: healthService,
	}
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
	response, isHealthy := hh.healthService.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if isHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		hh.logger.ErrorArgs(r.Context(), "Failed to encode health response: %v", err)
	}
}
