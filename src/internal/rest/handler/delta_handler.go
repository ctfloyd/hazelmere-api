package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/middleware"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/go-chi/chi/v5"
	chiWare "github.com/go-chi/chi/v5/middleware"
)

type DeltaHandler struct {
	monitor *monitor.Monitor
	service delta.DeltaService
}

func NewDeltaHandler(mon *monitor.Monitor, service delta.DeltaService) *DeltaHandler {
	return &DeltaHandler{mon, service}
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
	ctx, span := dh.monitor.StartSpan(r.Context(), "DeltaHandler.GetLatestDelta")
	defer span.End()

	userId := chi.URLParam(r, "userId")
	dh.monitor.Logger().InfoArgs(ctx, "Getting latest delta for user: %s", userId)

	d, err := dh.service.GetLatestDeltaForUser(ctx, userId)
	if err != nil {
		if errors.Is(err, delta.ErrDeltaNotFound) {
			dh.monitor.Logger().WarnArgs(ctx, "Delta not found for user %s", userId)
			hz_handler.Error(w, service_error.DeltaNotFound, "Delta not found.")
			return
		}
		dh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while getting latest delta for user %s: %+v", userId, err)
		hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting latest delta.")
		return
	}

	response := api.GetLatestDeltaResponse{
		Delta: d.ToAPI(),
	}

	hz_handler.Ok(w, response)
}

func (dh *DeltaHandler) GetDeltaInterval(w http.ResponseWriter, r *http.Request) {
	ctx, span := dh.monitor.StartSpan(r.Context(), "DeltaHandler.GetDeltaInterval")
	defer span.End()

	var intervalRequest api.GetDeltaIntervalRequest
	if ok := hz_handler.ReadBody(w, r, &intervalRequest); !ok {
		return
	}

	dh.monitor.Logger().InfoArgs(ctx, "Getting delta interval: %v", intervalRequest)
	result, err := dh.service.GetDeltasInRange(ctx, intervalRequest.UserId, intervalRequest.StartTime, intervalRequest.EndTime)
	if err != nil {
		if errors.Is(err, delta.ErrInvalidDeltaRequest) {
			dh.monitor.Logger().WarnArgs(ctx, "Invalid delta interval request: %+v", err)
			hz_handler.Error(w, service_error.BadRequest, err.Error())
		} else {
			dh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while getting delta interval: %+v", err)
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
	ctx, span := dh.monitor.StartSpan(r.Context(), "DeltaHandler.GetDeltaSummary")
	defer span.End()

	var summaryRequest api.GetDeltaSummaryRequest
	if ok := hz_handler.ReadBody(w, r, &summaryRequest); !ok {
		return
	}

	dh.monitor.Logger().InfoArgs(ctx, "Getting delta summary: %v", summaryRequest)
	summary, err := dh.service.GetDeltaSummary(ctx, summaryRequest.UserId, summaryRequest.StartTime, summaryRequest.EndTime)
	if err != nil {
		if errors.Is(err, delta.ErrInvalidDeltaRequest) {
			dh.monitor.Logger().WarnArgs(ctx, "Invalid delta summary request: %+v", err)
			hz_handler.Error(w, service_error.BadRequest, err.Error())
		} else {
			dh.monitor.Logger().ErrorArgs(ctx, "An unexpected error occurred while getting delta summary: %+v", err)
			hz_handler.Error(w, service_error.Internal, "An unexpected error occurred while getting delta summary.")
		}
		return
	}

	hz_handler.Ok(w, summary)
}
