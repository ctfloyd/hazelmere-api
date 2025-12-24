package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"github.com/ctfloyd/hazelmere-api/src/internal/rest/service_error"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_handler"
)

type Authorizer struct {
	enabled       bool
	allowedTokens []string
	monitor       *monitor.Monitor
}

func NewAuthorizer(enabled bool, allowedTokens []string, mon *monitor.Monitor) *Authorizer {
	return &Authorizer{enabled, allowedTokens, mon}
}

func (a *Authorizer) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.monitor.StartSpan(r.Context(), "Authorizer.Authorize")
		defer span.End()

		headerToken := r.Header.Get("Authorization")
		if !a.enabled {
			a.monitor.Logger().Info(ctx, "Authorization header present but auth is disabled.")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if headerToken == "" {
			a.monitor.Logger().Warn(ctx, "No authorization header.")
			unauthorized(w)
			return
		}

		parts := strings.Split(headerToken, " ")
		if len(parts) != 2 {
			a.monitor.Logger().Warn(ctx, "Malformed authorization header.")
			unauthorized(w)
			return
		}

		token := parts[1]
		if !slices.Contains(a.allowedTokens, token) {
			a.monitor.Logger().Warn(ctx, "Token not allowed.")
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter) {
	hz_handler.Error(w, service_error.Unauthorized, "You are not permitted to access this resource.")
}
