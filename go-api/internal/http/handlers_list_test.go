package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"fullstack-oracle/go-api/internal/domain"
)

type fakeSvcAll struct {
	out   []domain.Item
	total int64
	err   error
}

func (f *fakeSvcAll) List(ctx context.Context, page, size int, sort, q string) ([]domain.Item, int64, error) {
	return f.out, f.total, f.err
}
func (f *fakeSvcAll) ListStamp(context.Context) (time.Time, int, error)  { return time.Now(), 0, nil }
func (f *fakeSvcAll) GetStamp(context.Context, int64) (time.Time, error) { return time.Now(), nil }
func (f *fakeSvcAll) Get(context.Context, int64) (domain.Item, error)    { return domain.Item{}, nil }
func (f *fakeSvcAll) Create(context.Context, domain.CreateItemDTO) (domain.Item, error) {
	return domain.Item{}, nil
}
func (f *fakeSvcAll) Update(context.Context, int64, domain.CreateItemDTO) (domain.Item, error) {
	return domain.Item{}, nil
}
func (f *fakeSvcAll) Delete(context.Context, int64) error       { return nil }
func (f *fakeSvcAll) DeleteBulk(context.Context, []int64) error { return nil }

func TestListItems_BadParamsOK(t *testing.T) {
	h := &Handlers{S: &fakeSvcAll{}}

	req := httptest.NewRequest("GET", "/items/?page=abc&size=-1&sort=foo", nil)
	w := httptest.NewRecorder()

	h.ListItems(w, req)
	if w.Code != 200 {
		t.Fatalf("want 200 got %d", w.Code)
	}
}
