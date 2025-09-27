package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Registry = prometheus.NewRegistry()

	ReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Requests total"},
		[]string{"method", "path", "status"},
	)
	ErrTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_errors_total", Help: "5xx & handled errors"},
		[]string{"path", "status"},
	)
	Duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_duration_seconds",
			Help:    "Request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	Cache304 = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_cache_304_total", Help: "304 Not Modified hits"},
		[]string{"path"},
	)
	RateDrops = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "ratelimit_dropped_total", Help: "Requests dropped by rate limiter"},
	)
)

func init() {
	Registry.MustRegister(ReqTotal, ErrTotal, Duration, Cache304, RateDrops)
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			Duration.WithLabelValues(r.Method, r.URL.Path).
				Observe(time.Since(start).Seconds())
			ReqTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(sw.status)).Inc()

			if sw.status >= 500 {
				ErrTotal.WithLabelValues(r.URL.Path, strconv.Itoa(sw.status)).Inc()
			}
			if sw.status == http.StatusNotModified {
				Cache304.WithLabelValues(r.URL.Path).Inc()
			}
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
