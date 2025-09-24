package http

import (
	stdhttp "net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
	reqTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "count",
	}, []string{"method", "path", "status"})
)

func Metrics() func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			start := time.Now()
			ww := &rw{ResponseWriter: w}
			next.ServeHTTP(ww, r)
			status := ww.status
			if status == 0 {
				status = 200
			}
			labels := prometheus.Labels{"method": r.Method, "path": r.URL.Path, "status": httpStatusLabel(status)}
			reqTotal.With(labels).Inc()
			reqDuration.With(labels).Observe(time.Since(start).Seconds())
		})
	}
}
func httpStatusLabel(s int) string { return stdhttp.StatusText(s) }

func MetricsHandler() stdhttp.Handler { return promhttp.Handler() }
