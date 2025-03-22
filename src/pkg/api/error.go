package api

import "time"

type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
