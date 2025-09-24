package repo

import (
	"context"
	"database/sql"
	"errors"

	"fullstack-oracle/go-api/internal/domain"
)

var ErrNotFound = errors.New("not found")

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

func (r *ItemRepo) Get(ctx context.Context, id int64) (domain.Item, error) {
	const q = `SELECT id,name,price,created_at FROM app.items WHERE id=$1`
	var it domain.Item
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Item{}, ErrNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (r *ItemRepo) Create(ctx context.Context, in domain.CreateItemDTO) (domain.Item, error) {
	const q = `INSERT INTO app.items(name,price) VALUES($1,$2) RETURNING id,name,price,created_at`
	var it domain.Item
	err := r.DB.QueryRowContext(ctx, q, in.Name, in.Price).
		Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt)
	if err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (r *ItemRepo) Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error) {
	const q = `UPDATE app.items SET name=$1, price=$2 WHERE id=$3
	           RETURNING id,name,price,created_at`
	var it domain.Item
	err := r.DB.QueryRowContext(ctx, q, in.Name, in.Price, id).
		Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Item{}, ErrNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (r *ItemRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM app.items WHERE id=$1`
	res, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}
