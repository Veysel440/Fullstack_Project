package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"fullstack-oracle/go-api/internal/domain"
)

var ErrNotFound = errors.New("not found")

type ItemRepo struct{ DB *sql.DB }

func NewItemRepo(db *sql.DB) *ItemRepo { return &ItemRepo{DB: db} }

func (r *ItemRepo) ListPagedSorted(ctx context.Context, limit, offset int, sort string) ([]domain.Item, error) {
	col, dir := "id", "DESC"
	if sort != "" {
		parts := strings.Split(sort, ",")
		if len(parts) > 0 {
			switch parts[0] {
			case "id", "name", "price", "created_at":
				col = parts[0]
			}
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "asc" {
			dir = "ASC"
		}
	}
	q := fmt.Sprintf(`SELECT id,name,price,created_at FROM app.items ORDER BY %s %s LIMIT $1 OFFSET $2`, col, dir)
	rows, err := r.DB.QueryContext(ctx, q, limit, offset)
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

func (r *ItemRepo) ListPaged(ctx context.Context, limit, offset int) ([]domain.Item, error) {
	const q = `SELECT id, name, price, created_at
	          FROM app.items
	          ORDER BY id DESC
	          LIMIT $1 OFFSET $2`
	rows, err := r.DB.QueryContext(ctx, q, limit, offset)
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
	if errors.Is(err, sql.ErrNoRows) {
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
	err := r.DB.QueryRowContext(ctx, q, in.Name, in.Price).Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt)
	if err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (r *ItemRepo) Update(ctx context.Context, id int64, in domain.CreateItemDTO) (domain.Item, error) {
	const q = `UPDATE app.items SET name=$1, price=$2 WHERE id=$3
	           RETURNING id,name,price,created_at`
	var it domain.Item
	err := r.DB.QueryRowContext(ctx, q, in.Name, in.Price, id).Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, ErrNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (r *ItemRepo) ListStamp(ctx context.Context) (time.Time, int, error) {
	const q = `SELECT COALESCE(MAX(updated_at), MAX(created_at)) AS lm, COUNT(*) FROM app.items`
	var lm sql.NullTime
	var c int
	if err := r.DB.QueryRowContext(ctx, q).Scan(&lm, &c); err != nil {
		return time.Time{}, 0, err
	}
	if !lm.Valid {
		return time.Time{}, c, nil
	}
	return lm.Time.UTC(), c, nil
}

func (r *ItemRepo) GetStamp(ctx context.Context, id int64) (time.Time, error) {
	const q = `SELECT COALESCE(updated_at, created_at) FROM app.items WHERE id=$1`
	var t time.Time
	if err := r.DB.QueryRowContext(ctx, q, id).Scan(&t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, err
	}
	return t.UTC(), nil
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
