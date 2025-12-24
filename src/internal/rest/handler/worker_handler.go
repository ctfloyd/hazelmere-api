package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/worker"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type WorkerHandler struct {
	monitor *monitor.Monitor
	service worker.WorkerService
}

func NewWorkerHandler(mon *monitor.Monitor, service worker.WorkerService) *WorkerHandler {
	return &WorkerHandler{mon, service}
}

func (wh *WorkerHandler) RegisterRoutes(mux *chi.Mux, version ApiVersion, authorizer *middleware.Authorizer) {
	if version == ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(10 * time.Second))
			r.Use(authorizer.Authorize)
			r.Get(fmt.Sprintf("/v1/worker/snapshot/on-demand/{userId:%s}", hz_handler.RegexUuid), wh.GenerateSnapshotOnDemand)
		})
	}
}

func (wh *WorkerHandler) GenerateSnapshotOnDemand(w http.ResponseWriter, r *http.Request) {
	ctx, span := wh.monitor.StartSpan(r.Context(), "WorkerHandler.GenerateSnapshotOnDemand")
	defer span.End()

	userId := chi.URLParam(r, "userId")

	wh.monitor.Logger().InfoArgs(ctx, "Generating snapshot on demand for user: %s", userId)

	ss, err := wh.service.GenerateSnapshotOnDemand(ctx, userId)
	if err != nil {
		if errors.Is(err, worker.ErrHiscoreTimeout) {
			wh.monitor.Logger().WarnArgs(ctx, "Hiscore timeout while generating snapshot for user: %s", userId)
			hz_handler.Error(w, service_error.HiscoreTimeout, "Osrs hiscores timed out.")
			return
		}
		wh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while generating snapshot for user %s: %+v", userId, err)
		hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the worker operation.")
		return
	}

	response := api.GenerateSnapshotOnDemandResponse{
		Snapshot: ss.ToAPI(),
	}

	hz_handler.Ok(w, response)
}
