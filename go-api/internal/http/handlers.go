package http

import (
	"encoding/json"
	"net/http"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/service"
)

type Handlers struct{ S *service.ItemService }

func (h *Handlers) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (h *Handlers) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.S.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(items)
}

func (h *Handlers) CreateItem(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	it, err := h.S.Create(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(it)
}
