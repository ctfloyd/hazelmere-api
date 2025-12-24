package middleware

import (
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"net/http"
	"slices"
	"strings"
)

type Authorizer struct {
	enabled       bool
	allowedTokens []string
	logger        hz_logger.Logger
}

func NewAuthorizer(enabled bool, allowedTokens []string, logger hz_logger.Logger) *Authorizer {
	return &Authorizer{enabled, allowedTokens, logger}
}

func (a *Authorizer) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerToken := r.Header.Get("Authorization")
		if !a.enabled {
			a.logger.Info(r.Context(), "Authorization header present but auth is disabled.")
			next.ServeHTTP(w, r)
			return
		}

		if headerToken == "" {
			a.logger.Warn(r.Context(), "No authorization header.")
			unauthorized(w)
			return
		}

		parts := strings.Split(headerToken, " ")
		if len(parts) != 2 {
			a.logger.Warn(r.Context(), "Malformed authorization header.")
			unauthorized(w)
			return
		}

		token := parts[1]
		if !slices.Contains(a.allowedTokens, token) {
			a.logger.Warn(r.Context(), "Token not allowed.")
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter) {
	hz_handler.Error(w, service_error.Unauthorized, "You are not permitted to access this resource.")
}
