package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/health"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type HealthHandler struct {
	monitor       *monitor.Monitor
	healthService *health.Service
}

func NewHealthHandler(mon *monitor.Monitor, healthService *health.Service) *HealthHandler {
	return &HealthHandler{
		monitor:       mon,
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
	ctx, span := hh.monitor.StartSpan(r.Context(), "HealthHandler.HealthCheck")
	defer span.End()

	response, isHealthy := hh.healthService.Check(ctx)

	w.Header().Set("Content-Type", "application/json")
	if isHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		hh.monitor.Logger().ErrorArgs(ctx, "Failed to encode health response: %v", err)
	}
}
