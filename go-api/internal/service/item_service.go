package service

import (
	"context"
	"time"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
)

type ItemService struct{ R *repo.ItemRepo }

func NewItemService(r *repo.ItemRepo) *ItemService { return &ItemService{R: r} }

func ctx5(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

func (s *ItemService) List(ctx context.Context) ([]domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.R.List(c)
}

func (s *ItemService) Get(ctx context.Context, id int64) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.R.Get(c, id)
}

func (s *ItemService) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.R.Create(c, in)
}

func (s *ItemService) Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.R.Update(c, id, in)
}

func (s *ItemService) Delete(ctx context.Context, id int64) error {
	c, cancel := ctx5(ctx)
	defer cancel()
	return s.R.Delete(c, id)
}
