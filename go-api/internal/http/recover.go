package http

import (
	"bytes"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"runtime"

	"github.com/getsentry/sentry-go"
)

func Recover(l *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					buf := make([]byte, 64<<10)
					n := runtime.Stack(buf, false)
					stack := string(bytes.TrimSpace(buf[:n]))
					l.Error("panic",
						"rid", RequestIDFrom(r.Context()),
						"panic", rec,
						"stack", stack,
					)
					if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
						hub.CaptureException(fmt.Errorf("panic: %v", rec))
					} else {
						sentry.CaptureException(fmt.Errorf("panic: %v", rec))
					}
					writeError(w, r, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
