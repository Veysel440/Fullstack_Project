// internal/metrics/metrics.go
package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var Registry = prometheus.NewRegistry()

var reqDur = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "http_request_duration_seconds", Buckets: prometheus.DefBuckets},
	[]string{"method", "path", "status"},
)

func init() { Registry.MustRegister(reqDur) }

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, code: 200}
			start := time.Now()
			next.ServeHTTP(sw, r)
			reqDur.WithLabelValues(r.Method, r.URL.Path, http.StatusText(sw.code)).
				Observe(time.Since(start).Seconds())
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
