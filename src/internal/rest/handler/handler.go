package handler

import (
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/middleware"
	"github.com/go-chi/chi/v5"
)

type ApiVersion int

const (
	_ ApiVersion = iota
	ApiVersionV1
)

type HazelmereHandler interface {
	RegisterRoutes(mux *chi.Mux, version ApiVersion, authorizer *middleware.Authorizer)
}
