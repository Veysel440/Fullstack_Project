package http

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"fullstack-oracle/go-api/internal/domain"
)

type fakeDeleter struct{ called bool }

func (f *fakeDeleter) DeleteBulk(ctx context.Context, ids []int64) error { f.called = true; return nil }
func (f *fakeDeleter) List(context.Context, int, int, string, string) ([]domain.Item, int64, error) {
	return nil, 0, nil
}
func (f *fakeDeleter) ListStamp(context.Context) (time.Time, int, error)  { return time.Now(), 0, nil }
func (f *fakeDeleter) GetStamp(context.Context, int64) (time.Time, error) { return time.Now(), nil }
func (f *fakeDeleter) Get(context.Context, int64) (domain.Item, error)    { return domain.Item{}, nil }
func (f *fakeDeleter) Create(context.Context, domain.CreateItemDTO) (domain.Item, error) {
	return domain.Item{}, nil
}
func (f *fakeDeleter) Update(context.Context, int64, domain.CreateItemDTO) (domain.Item, error) {
	return domain.Item{}, nil
}
func (f *fakeDeleter) Delete(context.Context, int64) error { return nil }

func TestBulkDelete_OK(t *testing.T) {
	svc := &fakeDeleter{}
	h := &Handlers{S: svc}
	
	req := httptest.NewRequest("DELETE", "/items/bulk", bytes.NewBufferString(`{"ids":[1,2,3]}`))
	w := httptest.NewRecorder()

	h.BulkDeleteItems(w, req)

	if w.Code != 204 {
		t.Fatalf("want 204 got %d", w.Code)
	}
	if !svc.called {
		t.Fatal("service not called")
	}
}
