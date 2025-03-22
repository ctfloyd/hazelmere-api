package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

type Middleware struct {
	logger Logger
}

func NewMiddleware(logger Logger) *Middleware {
	return &Middleware{logger: logger}
}

func (mw *Middleware) Serve(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			mw.logger.InfoArgs(r.Context(), "%s %s - %d, %d bytes written in %s", r.Method, r.URL, ww.Status(), ww.BytesWritten(), time.Since(t1))
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
