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

func normalizeSort(in string) (col, dir string) {
	col, dir = "id", "DESC"
	if in == "" {
		return
	}
	p := strings.Split(in, ",")
	if len(p) > 0 {
		switch p[0] {
		case "id", "name", "price", "created_at":
			col = p[0]
		}
	}
	if len(p) > 1 && strings.EqualFold(p[1], "asc") {
		dir = "ASC"
	}
	return
}

func (r *ItemRepo) ListPagedSortedWithTotal(ctx context.Context, limit, offset int, sort, q string) ([]domain.Item, int64, error) {
	col, dir := normalizeSort(sort)

	var where string
	var args []any
	argN := 0
	if q != "" {
		argN++
		where = fmt.Sprintf("WHERE name ILIKE $%d", argN)
		args = append(args, "%"+q+"%")
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM app.items " + where
	if err := r.DB.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argN++
	args = append(args, limit)
	limitPos := argN

	argN++
	args = append(args, offset)
	offsetPos := argN

	listSQL := fmt.Sprintf(`
		SELECT id,name,price,created_at
		  FROM app.items
		  %s
		  ORDER BY %s %s
		  LIMIT $%d OFFSET $%d
	`, where, col, dir, limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Item
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(&it.ID, &it.Name, &it.Price, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
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
	const q = `INSERT INTO app.items(name,price) VALUES($1,$2)
	           RETURNING id,name,price,created_at`
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
	if errors.Is(err, sql.ErrNoRows) {
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

func (r *ItemRepo) ListStamp(ctx context.Context) (time.Time, int, error) {
	const q = `SELECT COALESCE(MAX(created_at), now()), COUNT(*) FROM app.items`
	var lm time.Time
	var c int
	if err := r.DB.QueryRowContext(ctx, q).Scan(&lm, &c); err != nil {
		return time.Time{}, 0, err
	}
	return lm.UTC(), c, nil
}

func (r *ItemRepo) GetStamp(ctx context.Context, id int64) (time.Time, error) {
	const q = `SELECT created_at FROM app.items WHERE id=$1`
	var t time.Time
	if err := r.DB.QueryRowContext(ctx, q, id).Scan(&t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, err
	}
	return t.UTC(), nil
}
