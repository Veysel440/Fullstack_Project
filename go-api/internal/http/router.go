package http

import (
	"log/slog"
	stdhttp "net/http"
	"os"

	"fullstack-oracle/go-api/internal/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Router(
	h *Handlers,
	corsMW func(stdhttp.Handler) stdhttp.Handler,
	logger *slog.Logger,
	rl *RateLimiter,
	jwtv *JWTVerifier,
	ah *AuthHandlers,
) stdhttp.Handler {
	r := chi.NewRouter()
	r.Use(RequestID())
	r.Use(Recover(logger))
	r.Use(Logger(logger))
	if rl != nil {
		r.Use(rl.Middleware())
	}
	r.Use(corsMW)

	r.Mount("/debug", middleware.Profiler())
	r.Get("/openapi.yaml", OpenAPISpec)
	r.Get("/docs", Docs)

	r.Use(metrics.Middleware())
	mh := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
	if u := os.Getenv("METRICS_USER"); u != "" {
		mh = BasicAuth(u, os.Getenv("METRICS_PASS"))(mh)
	}
	r.Method("GET", "/metrics", mh)

	r.Get("/health", h.Health)

	if ah != nil {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", ah.Login)
			r.Post("/refresh", ah.Refresh)
		})
	}

	if jwtv != nil && ah != nil {
		r.Group(func(pr chi.Router) {
			pr.Use(jwtv.AuthRequired("user", "admin"))
			pr.Get("/auth/me", ah.Me)

			pr.Route("/items", func(r chi.Router) {
				r.Get("/", h.ListItems)
				r.Post("/", h.CreateItem)
				r.Get("/{id}", h.GetItem)
				r.Put("/{id}", h.UpdateItem)
				r.With(jwtv.AuthRequired("admin")).Delete("/{id}", h.DeleteItem)
				r.With(jwtv.AuthRequired("admin")).Delete("/bulk", h.BulkDeleteItems)
			})
		})
	}

	return r
}
