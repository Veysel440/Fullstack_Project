package repo

import (
	"context"
	"database/sql"

	"fullstack-oracle/go-api/internal/domain"
)

type ItemRepo struct{ DB *sql.DB }

func NewItemRepo(db *sql.DB) *ItemRepo { return &ItemRepo{DB: db} }

func (r *ItemRepo) List(ctx context.Context) ([]domain.Item, error) {
	const q = `SELECT id,name,price,created_at FROM app.items ORDER BY id DESC LIMIT 100`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Item
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *ItemRepo) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	const q = `INSERT INTO app.items(name,price) VALUES($1,$2) RETURNING id,created_at`
	var it domain.Item
	it.Name, it.Price = in.Name, in.Price
	if err := r.DB.QueryRowContext(ctx, q, in.Name, in.Price).Scan(&it.ID, &it.CreatedAt); err != nil {
		return domain.Item{}, err
	}
	return it, nil
}
