package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handlers struct{ S *service.ItemService }

var v = validator.New()

func (h *Handlers) Health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{"ok": true})
}

func (h *Handlers) ListItems(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	items, err := h.S.List(r.Context(), page, size)
	if err != nil {
		writeError(w, r, 500, "list_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusOK, items)
}

func (h *Handlers) GetItem(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeValidation(w, r, map[string]string{"id": "must be integer"})
		return
	}
	it, err := h.S.Get(r.Context(), id)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, r, 404, "not_found", "item not found")
		return
	}
	if err != nil {
		writeError(w, r, 500, "get_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusOK, it)
}

func (h *Handlers) CreateItem(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var dto domain.CreateItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := v.Struct(dto); err != nil {
		writeValidation(w, r, toFields(err))
		return
	}
	it, err := h.S.Create(r.Context(), dto)
	if err != nil {
		writeError(w, r, 500, "create_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusCreated, it)
}

func (h *Handlers) UpdateItem(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeValidation(w, r, map[string]string{"id": "must be integer"})
		return
	}
	var dto domain.CreateItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := v.Struct(dto); err != nil {
		writeValidation(w, r, toFields(err))
		return
	}
	it, err := h.S.Update(r.Context(), id, dto)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, r, 404, "not_found", "item not found")
		return
	}
	if err != nil {
		writeError(w, r, 500, "update_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusOK, it)
}

func (h *Handlers) DeleteItem(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeValidation(w, r, map[string]string{"id": "must be integer"})
		return
	}
	if err := h.S.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeError(w, r, 404, "not_found", "item not found")
			return
		}
		writeError(w, r, 500, "delete_failed", err.Error())
		return
	}
	w.WriteHeader(stdhttp.StatusNoContent)
}

func parseID(v string) (int64, error) { return strconv.ParseInt(v, 10, 64) }

func toFields(err error) map[string]string {
	out := map[string]string{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			out[strings.ToLower(fe.Field())] = fe.Tag()
		}
	}
	return out
}
