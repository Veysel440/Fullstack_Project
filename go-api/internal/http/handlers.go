package http

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handlers struct{ S *service.ItemService }

var v = validator.New()

type ItemPort interface {
	List(ctx context.Context, page, size int, sort, q string) ([]domain.Item, int64, error)
	ListStamp(ctx context.Context) (time.Time, int, error)
	GetStamp(ctx context.Context, id int64) (time.Time, error)

	Get(ctx context.Context, id int64) (domain.Item, error)
	Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error)
	Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error)
	Delete(ctx context.Context, id int64) error

	// bulk
	DeleteBulk(ctx context.Context, ids []int64) error
}

type PagedItems struct {
	Items []domain.Item `json:"items"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
	Total int64         `json:"total"`
}

func (h *Handlers) ListItems(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	sort := r.URL.Query().Get("sort")
	q := r.URL.Query().Get("q")

	items, total, err := h.S.List(r.Context(), page, size, sort, q)
	if err != nil {
		writeError(w, r, 500, "list_failed", err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	writeJSON(w, stdhttp.StatusOK, PagedItems{
		Items: items, Page: page, Size: size, Total: total,
	})
}

func (h *Handlers) Health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{"ok": true})
}

func weakTag(ts time.Time, n int) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%d-%d", ts.Unix(), n)))
	return `W/"` + fmt.Sprintf("%x", h[:]) + `"`
}

func (h *Handlers) GetItem(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeValidation(w, r, map[string]string{"id": "must be integer"})
		return
	}

	if ts, e := h.S.GetStamp(r.Context(), id); e == nil {
		tag := weakTag(ts, int(id))
		if r.Header.Get("If-None-Match") == tag {
			w.WriteHeader(stdhttp.StatusNotModified)
			return
		}
		w.Header().Set("ETag", tag)
		w.Header().Set("Cache-Control", "private, max-age=60")
	}

	it, err := h.S.Get(r.Context(), id)
	if err == repo.ErrNotFound {
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

func (h *Handlers) BulkDeleteItems(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var ids []int64
	var probe any
	if err := json.NewDecoder(r.Body).Decode(&probe); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	switch t := probe.(type) {
	case []any:
		for _, v := range t {
			switch n := v.(type) {
			case float64:
				ids = append(ids, int64(n))
			}
		}
	case map[string]any:
		if arr, ok := t["ids"].([]any); ok {
			for _, v := range arr {
				if n, ok2 := v.(float64); ok2 {
					ids = append(ids, int64(n))
				}
			}
		}
	default:
		writeValidation(w, r, map[string]string{"ids": "required"})
		return
	}
	if len(ids) == 0 {
		writeValidation(w, r, map[string]string{"ids": "empty"})
		return
	}
	if err := h.S.DeleteBulk(r.Context(), ids); errors.Is(err, repo.ErrNotFound) {
		writeError(w, r, 404, "not_found", "no items deleted")
		return
	} else if err != nil {
		writeError(w, r, 500, "bulk_delete_failed", err.Error())
		return
	}
	w.WriteHeader(stdhttp.StatusNoContent)
}
