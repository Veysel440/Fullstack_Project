package service

import (
	"context"
	"encoding/json"
	"time"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/events"
	"fullstack-oracle/go-api/internal/repo"
)

type ItemService struct {
	r  *repo.ItemRepo
	ev *events.Writer
}

func NewItemService(r *repo.ItemRepo, ev *events.Writer) *ItemService {
	return &ItemService{r: r, ev: ev}
}

func ctx5(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

func (s *ItemService) List(ctx context.Context, page, size int, sort, q string) ([]domain.Item, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.r.ListPagedSortedWithTotal(ctx, size, (page-1)*size, sort, q)
}

func (s *ItemService) ListStamp(ctx context.Context) (time.Time, int, error) {
	return s.r.ListStamp(ctx)
}
func (s *ItemService) GetStamp(ctx context.Context, id int64) (time.Time, error) {
	return s.r.GetStamp(ctx, id)
}

func (s *ItemService) Get(ctx context.Context, id int64) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.r.Get(c, id)
}

func (s *ItemService) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	it, err := s.r.Create(c, in)
	if err == nil && s.ev != nil {
		b, _ := json.Marshal(map[string]any{"type": "item.created", "item": it})
		_ = s.ev.Publish(ctx, "item", b)
	}
	return it, err
}

func (s *ItemService) Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	it, err := s.r.Update(c, id, in)
	if err == nil && s.ev != nil {
		b, _ := json.Marshal(map[string]any{"type": "item.updated", "item": it})
		_ = s.ev.Publish(ctx, "item", b)
	}
	return it, err
}

func (s *ItemService) Delete(ctx context.Context, id int64) error {
	c, cancel := ctx5(ctx)
	defer cancel()
	err := s.r.Delete(c, id)
	if err == nil && s.ev != nil {
		b, _ := json.Marshal(map[string]any{"type": "item.deleted", "id": id})
		_ = s.ev.Publish(ctx, "item", b)
	}
	return err
}
