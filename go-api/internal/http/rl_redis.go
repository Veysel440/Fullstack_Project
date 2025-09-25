package http

import (
	"net"
	"net/http"

	"fullstack-oracle/go-api/internal/cache"
)

func RedisRateLimit(l *cache.Limiter) func(http.Handler) http.Handler {
	if l == nil {
		return func(h http.Handler) http.Handler { return h }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !l.Allow(r.Context(), ip) {
				writeError(w, r, 429, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
