package snapshot

import (
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/handler"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/logger"
	"github.com/ctfloyd/hazelmere-api/src/internal/common/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
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
		mux.Get(fmt.Sprintf("/v1/snapshot/{userId:%s}/nearest/{timestamp}", handler.RegexUuid), sh.GetSnapshotForUserNearestTimestamp)
		mux.Post("/v1/snapshot", sh.CreateSnapshot)
	}
}

func (sh *SnapshotHandler) GetAllSnapshotsForUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	sh.logger.InfoArgs(r.Context(), "Getting all snapshots for user: %s", userId)

	snapshots, err := sh.service.GetAllSnapshotsForUser(r.Context(), userId)
	if err != nil {
		handler.Error(w, service_error.Internal, "An unexpected error occurred while getting all snapshots for user.")
	}

	response := api.GetAllSnapshotsForUser{
		Snapshots: MapManyDomainToApi(snapshots),
	}

	handler.Ok(w, response)
}

func (sh *SnapshotHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var createSnapshotRequest api.CreateSnapshotRequest
	if ok := handler.ReadBody(w, r, &createSnapshotRequest); !ok {
		return
	}

	snapshot, err := sh.service.CreateSnapshot(r.Context(), MapApiToDomain(createSnapshotRequest.Snapshot))
	if err != nil {
		if errors.Is(err, ErrSnapshotValidation) {
			handler.Error(w, service_error.InvalidSnapshot, err.Error())
			return
		}

		handler.Error(w, service_error.Internal, "An unexpected error occurred while creating snapshot.")
		return
	}

	response := api.CreateSnapshotResponse{
		Snapshot: MapDomainToApi(snapshot),
	}

	handler.Ok(w, response)
}

func (sh *SnapshotHandler) GetSnapshotForUserNearestTimestamp(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")

	timestampString := chi.URLParam(r, "timestamp")
	millis, err := strconv.ParseInt(timestampString, 10, 64)
	if err != nil {
		handler.Error(w, service_error.BadRequest, "Could not convert timestamp to a number.")
		return
	}

	sh.logger.InfoArgs(r.Context(), "Getting snapshots for user: %s closest to %d", userId, millis)

	snapshot, err := sh.service.GetSnapshotForUserNearestTimestamp(r.Context(), userId, millis)
	if err != nil {
		if errors.Is(err, ErrSnapshotNotFound) {
			handler.Error(w, service_error.SnapshotNotFound, "Snapshot not found.")
			return
		}
		handler.Error(w, service_error.Internal, "An unexpected error occurred while getting snapshots for user nearest timestamp.")
		return
	}

	response := api.GetSnapshotNearestTimestampResponse{
		Snapshot: MapDomainToApi(snapshot),
	}

	handler.Ok(w, response)
}
