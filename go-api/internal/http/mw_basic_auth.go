package http

import (
	"crypto/subtle"
	"net/http"
)

func BasicAuth(user, pass string) func(http.Handler) http.Handler {
	if user == "" {
		return func(h http.Handler) http.Handler { return h }
	}
	uu := []byte(user)
	pp := []byte(pass)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(u), uu) != 1 ||
				subtle.ConstantTimeCompare([]byte(p), pp) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
