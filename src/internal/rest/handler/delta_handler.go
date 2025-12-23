package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type DeltaHandler struct {
	logger  hz_logger.Logger
	service delta.DeltaService
}

func NewDeltaHandler(logger hz_logger.Logger, service delta.DeltaService) *DeltaHandler {
	return &DeltaHandler{logger, service}
}

func (dh *DeltaHandler) RegisterRoutes(mux *chi.Mux, version ApiVersion, authorizer *middleware.Authorizer) {
	if version == ApiVersionV1 {
		mux.Group(func(r chi.Router) {
			r.Use(chiWare.Timeout(5000 * time.Millisecond))
			r.Get(fmt.Sprintf("/v1/delta/{userId:%s}/latest", hz_handler.RegexUuid), dh.GetLatestDelta)
			r.Post("/v1/delta/interval", dh.GetDeltaInterval)
			r.Post("/v1/delta/summary", dh.GetDeltaSummary)
		})
	}
}

func (dh *DeltaHandler) GetLatestDelta(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	dh.logger.InfoArgs(r.Context(), "Getting latest delta for user: %s", userId)

	d, err := dh.service.GetLatestDeltaForUser(r.Context(), userId)
	if err != nil {
		if errors.Is(err, delta.ErrDeltaNotFound) {
			dh.logger.WarnArgs(r.Context(), "Delta not found for user %s", userId)
			hz_handler.Error(w, service_error.DeltaNotFound, "Delta not found.")
			return
		}
		dh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while getting latest delta for user %s: %+v", userId, err)
		hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting latest delta.")
		return
	}

	response := api.GetLatestDeltaResponse{
		Delta: d.ToAPI(),
	}

	hz_handler.Ok(w, response)
}

func (dh *DeltaHandler) GetDeltaInterval(w http.ResponseWriter, r *http.Request) {
	var intervalRequest api.GetDeltaIntervalRequest
	if ok := hz_handler.ReadBody(w, r, &intervalRequest); !ok {
		return
	}

	dh.logger.InfoArgs(r.Context(), "Getting delta interval: %v", intervalRequest)
	result, err := dh.service.GetDeltasInRange(r.Context(), intervalRequest.UserId, intervalRequest.StartTime, intervalRequest.EndTime)
	if err != nil {
		if errors.Is(err, delta.ErrInvalidDeltaRequest) {
			dh.logger.WarnArgs(r.Context(), "Invalid delta interval request: %+v", err)
			hz_handler.Error(w, service_error.BadRequest, err.Error())
		} else {
			dh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while getting delta interval: %+v", err)
			hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting delta interval.")
		}
		return
	}

	response := api.GetDeltaIntervalResponse{
		Deltas:      delta.HiscoreDelta{}.ManyToAPI(result.Deltas),
		TotalDeltas: result.TotalDeltas,
	}

	hz_handler.Ok(w, response)
}

func (dh *DeltaHandler) GetDeltaSummary(w http.ResponseWriter, r *http.Request) {
	var summaryRequest api.GetDeltaSummaryRequest
	if ok := hz_handler.ReadBody(w, r, &summaryRequest); !ok {
		return
	}

	dh.logger.InfoArgs(r.Context(), "Getting delta summary: %v", summaryRequest)
	summary, err := dh.service.GetDeltaSummary(r.Context(), summaryRequest.UserId, summaryRequest.StartTime, summaryRequest.EndTime)
	if err != nil {
		if errors.Is(err, delta.ErrInvalidDeltaRequest) {
			dh.logger.WarnArgs(r.Context(), "Invalid delta summary request: %+v", err)
			hz_handler.Error(w, service_error.BadRequest, err.Error())
		} else {
			dh.logger.ErrorArgs(r.Context(), "An unexpected error occurred while getting delta summary: %+v", err)
			hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting delta summary.")
		}
		return
	}

	hz_handler.Ok(w, summary)
}
