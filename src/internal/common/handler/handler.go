package handler

import (
	"api/src/internal/common/service_error"
	"api/src/pkg/api"
	"fmt"
	"github.com/go-chi/chi/v5"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"time"
)

type ApiVersion int

const (
	_ ApiVersion = iota
	ApiVersionV1
)

const RegexUuid string = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`

type HazelmereHandler interface {
	RegisterRoutes(mux *chi.Mux, version ApiVersion)
}

func Error(w http.ResponseWriter, serviceError service_error.ServiceError, message string) {
	response := convertServiceErrorToResponse(serviceError, message)
	b, err := jsoniter.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	_, _ = w.Write(b)
}

func Ok(w http.ResponseWriter, response any) {
	b, err := jsoniter.Marshal(response)
	if err != nil {
		Error(w, service_error.Internal, "Could not marshal response.")
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		Error(w, service_error.Internal, "Could not write all bytes in the response.")
	}
}

func ReadBody(w http.ResponseWriter, r *http.Request, body any) bool {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, service_error.Internal, "An unexpected error occurred while reading request body.")
		return false
	}

	err = jsoniter.Unmarshal(bytes, body)
	if err != nil {
		Error(w, service_error.BadRequest, fmt.Sprintf("The request body could not be parsed. %v", err))
		return false
	}

	return true
}

func convertServiceErrorToResponse(serviceError service_error.ServiceError, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Code:      serviceError.Code,
		Status:    serviceError.Status,
		Message:   message,
		Timestamp: time.Now(),
	}
}
