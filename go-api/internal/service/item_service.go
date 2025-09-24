package service

import (
	"context"
	"time"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
)

type ItemService struct{ r *repo.ItemRepo }

func NewItemService(r *repo.ItemRepo) *ItemService { return &ItemService{r: r} }

func ctx5(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

func (s *ItemService) List(ctx context.Context, page, size int) ([]domain.Item, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.r.ListPaged(ctx, size, (page-1)*size)
}

func (s *ItemService) Get(ctx context.Context, id int64) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.r.Get(c, id)
}

func (s *ItemService) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.r.Create(c, in)
}

func (s *ItemService) Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.r.Update(c, id, in)
}

func (s *ItemService) Delete(ctx context.Context, id int64) error {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.r.Delete(c, id)
}
