package service_error

import "net/http"

type ServiceError struct {
	Code   string
	Status int
}

const (
	ErrorCodeInternal = "INTERNAL_SERVICE_ERROR"
)

var Internal = ServiceError{Code: ErrorCodeInternal, Status: http.StatusInternalServerError}
