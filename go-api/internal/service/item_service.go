package service

import (
	"context"
	"time"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
)

type ItemService struct{ R *repo.ItemRepo }

func NewItemService(r *repo.ItemRepo) *ItemService { return &ItemService{R: r} }

func (s *ItemService) List(ctx context.Context) ([]domain.Item, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.R.List(c)
}
func (s *ItemService) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.R.Create(c, in)
}
