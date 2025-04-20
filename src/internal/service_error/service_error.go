package service_error

import (
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_service_error"
	"net/http"
)

var Internal = hz_service_error.ServiceError{Code: api.ErrorCodeInternal, Status: http.StatusInternalServerError}
var BadRequest = hz_service_error.ServiceError{Code: api.ErrorCodeBadRequest, Status: http.StatusBadRequest}
var InvalidSnapshot = hz_service_error.ServiceError{Code: api.ErrorCodeInvalidSnapshot, Status: http.StatusBadRequest}
var SnapshotNotFound = hz_service_error.ServiceError{Code: api.ErrorCodeSnapshotNotFound, Status: http.StatusNotFound}
var UserNotFound = hz_service_error.ServiceError{Code: api.ErrorCodeUserNotFound, Status: http.StatusNotFound}
var RunescapeNameAlreadyTracked = hz_service_error.ServiceError{Code: api.ErrorCodeRunescapeNameAlreadyTracked, Status: http.StatusBadRequest}
var HiscoreTimeout = hz_service_error.ServiceError{Code: api.ErrorCodeHiscoreTimeout, Status: http.StatusRequestTimeout}
var Unauthorized = hz_service_error.ServiceError{Code: api.ErrorCodeUnauthorized, Status: http.StatusUnauthorized}
