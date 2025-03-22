package snapshot

import (
	"api/src/internal/common/handler"
	"api/src/internal/common/logger"
	"api/src/internal/common/service_error"
	"api/src/pkg/api"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type SnapshotHandler struct {
	logger  logger.Logger
	service SnapshotService
}

func NewSnapshotHandler(logger logger.Logger, service SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{logger, service}
}

func (sh *SnapshotHandler) RegisterRoutes(mux *chi.Mux, version handler.ApiVersion) {
	if version == handler.ApiVersionV1 {
		mux.Get(fmt.Sprintf("/v1/snapshot/{userId:%s}", handler.RegexUuid), sh.GetAllSnapshotsForUser)
		mux.Get("/v1/hello", sh.Hello)
	}
}

func (sh *SnapshotHandler) GetAllSnapshotsForUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	sh.logger.InfoArgs(r.Context(), "Getting all snapshots for user: %s", userId)

	snapshots, err := sh.service.GetAllSnapshotsForUser(r.Context(), userId)
	if err != nil {
		handler.Error(w, service_error.Internal, "An unexpected error occurred while getting all snapshots for user.")
	}

	response := api.GetAllHiscoreSnapshotsForUserResponse{
		Snapshots: MapManyDomainToApi(snapshots),
	}

	handler.Ok(w, response)
}

func (sh *SnapshotHandler) Hello(w http.ResponseWriter, r *http.Request) {
	handler.Ok(w, map[string]string{"hello": "world"})
}
