package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(h *Handlers, corsMW func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(corsMW)

	r.Get("/health", h.Health)
	r.Route("/items", func(r chi.Router) {
		r.Get("/", h.ListItems)
		r.Post("/", h.CreateItem)
	})
	return r
}
