package worker

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/service_error"
	"github.com/ctfloyd/hazelmere-api/src/internal/snapshot"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

type WorkerHandler struct {
	logger  hz_logger.Logger
	service WorkerService
}

func NewWorkerHandler(logger hz_logger.Logger, service WorkerService) *WorkerHandler {
	return &WorkerHandler{logger, service}
}

func (wh *WorkerHandler) RegisterRoutes(mux *chi.Mux, version handler.ApiVersion, authorizer *middleware.Authorizer) {
	if version == handler.ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(10 * time.Second))
			r.Use(authorizer.Authorize)
			r.Get(fmt.Sprintf("/v1/worker/snapshot/on-demand/{userId:%s}", hz_handler.RegexUuid), wh.GenerateSnapshotOnDemand)
		})
	}
}

func (wh *WorkerHandler) GenerateSnapshotOnDemand(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	ctx := r.Context()

	ss, err := wh.service.GenerateSnapshotOnDemand(ctx, userId)
	if err != nil {
		if errors.Is(err, ErrHiscoreTimeout) {
			hz_handler.Error(w, service_error.HiscoreTimeout, "Osrs hiscores timed out.")
		}
		wh.logger.ErrorArgs(ctx, "An unexpected error occurred while performing the worker operation: %v", err)
		hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while performing the worker operation.")
		return
	}

	response := api.GenerateSnapshotOnDemandResponse{
		Snapshot: snapshot.MapDomainToApi(ss),
	}

	hz_handler.Ok(w, response)
}
