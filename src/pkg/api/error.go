package api

import "time"

const (
	ErrorCodeInternal         = "INTERNAL_SERVICE_ERROR"
	ErrorCodeBadRequest       = "BAD_REQUEST"
	ErrorCodeInvalidSnapshot  = "INVALID_SNAPSHOT"
	ErrorCodeSnapshotNotFound = "SNAPSHOT_NOT_FOUND"
)

type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
