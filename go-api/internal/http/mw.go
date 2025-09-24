package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/google/uuid"
)

func RequestID() func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = uuid.NewString()
			}
			r = r.WithContext(withRequestID(r.Context(), id))
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r)
		})
	}
}

type rw struct {
	stdhttp.ResponseWriter
	status int
	size   int
}

func (w *rw) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }
func (w *rw) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = stdhttp.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func Logger(l *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			ww := &rw{ResponseWriter: w}
			next.ServeHTTP(ww, r)
			l.Info("http",
				"ts", time.Now().UTC().Format(time.RFC3339Nano),
				"rid", RequestIDFrom(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"size", ww.size,
				"dur_ms", time.Since(start).Milliseconds(),
				"ip", r.RemoteAddr,
				"ua", r.UserAgent(),
			)
		})
	}
}
