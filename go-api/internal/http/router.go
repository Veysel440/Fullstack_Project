package http

import (
	"log/slog"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(h *Handlers, corsMW func(stdhttp.Handler) stdhttp.Handler, logger *slog.Logger, rl *RateLimiter,
	jwtv *JWTVerifier, ah *AuthHandlers,
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

	r.Use(Metrics())
	r.Method(stdhttp.MethodGet, "/metrics", MetricsHandler())

	r.Get("/health", h.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", ah.Login)
		r.Post("/refresh", ah.Refresh)
	})

	r.Group(func(pr chi.Router) {
		pr.Use(jwtv.AuthRequired("user", "admin"))
		pr.Get("/auth/me", ah.Me)

		pr.Route("/items", func(r chi.Router) {
			r.Get("/", h.ListItems)
			r.Post("/", h.CreateItem)
			r.Get("/{id}", h.GetItem)
			r.Put("/{id}", h.UpdateItem)
			r.With(jwtv.AuthRequired("admin")).Delete("/{id}", h.DeleteItem)
		})
	})

	return r
}
