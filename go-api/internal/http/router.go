package http

import (
	"log/slog"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
)

func Router(h *Handlers, corsMW func(stdhttp.Handler) stdhttp.Handler, logger *slog.Logger, rl *RateLimiter) stdhttp.Handler {
	r := chi.NewRouter()
	r.Use(RequestID())
	r.Use(Recover(logger))
	r.Use(Logger(logger))
	if rl != nil {
		r.Use(rl.Middleware())
	}
	r.Use(corsMW)

	r.Get("/health", h.Health)
	r.Route("/items", func(r chi.Router) {
		r.Get("/", h.ListItems)
		r.Post("/", h.CreateItem)
		r.Get("/{id}", h.GetItem)
		r.Put("/{id}", h.UpdateItem)
		r.Delete("/{id}", h.DeleteItem)
	})
	return r
}
