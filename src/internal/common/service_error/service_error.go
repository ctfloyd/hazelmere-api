package service_error

import (
	"api/src/pkg/api"
	"net/http"
)

type ServiceError struct {
	Code   string
	Status int
}

var Internal = ServiceError{Code: api.ErrorCodeInternal, Status: http.StatusInternalServerError}
var BadRequest = ServiceError{Code: api.ErrorCodeBadRequest, Status: http.StatusBadRequest}
var InvalidSnapshot = ServiceError{Code: api.ErrorCodeInvalidSnapshot, Status: http.StatusBadRequest}
var SnapshotNotFound = ServiceError{Code: api.ErrorCodeSnapshotNotFound, Status: http.StatusNotFound}
