package snapshot

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strconv"
	"time"
)

type SnapshotHandler struct {
	logger  hz_logger.Logger
	service SnapshotService
}

func NewSnapshotHandler(logger hz_logger.Logger, service SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{logger, service}
}

func (sh *SnapshotHandler) RegisterRoutes(mux *chi.Mux, version handler.ApiVersion, authorizer *middleware.Authorizer) {
	if version == handler.ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(5000 * time.Millisecond))
			r.Get(fmt.Sprintf("/v1/snapshot/{userId:%s}/nearest/{timestamp}", hz_handler.RegexUuid), sh.GetSnapshotForUserNearestTimestamp)
			r.Group(func(secure chi.Router) {
				secure.Use(authorizer.Authorize)
				secure.Get(fmt.Sprintf("/v1/snapshot/{userId:%s}", hz_handler.RegexUuid), sh.GetAllSnapshotsForUser)
				secure.Post("/v1/snapshot", sh.CreateSnapshot)
			})
		})
	}
}

func (sh *SnapshotHandler) GetAllSnapshotsForUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	sh.logger.InfoArgs(r.Context(), "Getting all snapshots for user: %s", userId)

	snapshots, err := sh.service.GetAllSnapshotsForUser(r.Context(), userId)
	if err != nil {
		hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting all snapshots for user.")
		return
	}

	response := api.GetAllSnapshotsForUser{
		Snapshots: MapManyDomainToApi(snapshots),
	}

	hz_handler.Ok(w, response)
}

func (sh *SnapshotHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var createSnapshotRequest api.CreateSnapshotRequest
	if ok := hz_handler.ReadBody(w, r, &createSnapshotRequest); !ok {
		return
	}

	snapshot, err := sh.service.CreateSnapshot(r.Context(), MapApiToDomain(createSnapshotRequest.Snapshot))
	if err != nil {
		if errors.Is(err, ErrSnapshotValidation) {
			hz_handler.Error(w, service_error.InvalidSnapshot, err.Error())
			return
		}

		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while creating snapshot.")
		return
	}

	response := api.CreateSnapshotResponse{
		Snapshot: MapDomainToApi(snapshot),
	}

	hz_handler.Ok(w, response)
}

func (sh *SnapshotHandler) GetSnapshotForUserNearestTimestamp(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")

	timestampString := chi.URLParam(r, "timestamp")
	millis, err := strconv.ParseInt(timestampString, 10, 64)
	if err != nil {
		hz_handler.Error(w, service_error.BadRequest, "Could not convert timestamp to a number.")
		return
	}

	sh.logger.InfoArgs(r.Context(), "Getting snapshots for user: %s closest to %d", userId, millis)

	snapshot, err := sh.service.GetSnapshotForUserNearestTimestamp(r.Context(), userId, millis)
	if err != nil {
		if errors.Is(err, ErrSnapshotNotFound) {
			hz_handler.Error(w, service_error.SnapshotNotFound, "Snapshot not found.")
			return
		}
		hz_handler.Error(w, service_error.Internal, "An unexpected service_error occurred while getting snapshots for user nearest timestamp.")
		return
	}

	response := api.GetSnapshotNearestTimestampResponse{
		Snapshot: MapDomainToApi(snapshot),
	}

	hz_handler.Ok(w, response)
}
