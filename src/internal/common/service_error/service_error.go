package service_error

import "net/http"

type ServiceError struct {
	Code   string
	Status int
}

const (
	ErrorCodeInternal        = "INTERNAL_SERVICE_ERROR"
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeInvalidSnapshot = "INVALID_SNAPSHOT"
)

var Internal = ServiceError{Code: ErrorCodeInternal, Status: http.StatusInternalServerError}
var BadRequest = ServiceError{Code: ErrorCodeBadRequest, Status: http.StatusBadRequest}
var InvalidSnapshot = ServiceError{Code: ErrorCodeInvalidSnapshot, Status: http.StatusBadRequest}
