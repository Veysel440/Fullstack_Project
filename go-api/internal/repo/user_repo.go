package repo

import (
	"context"
	"database/sql"
	"errors"

	"fullstack-oracle/go-api/internal/domain"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	const q = `SELECT id, email, role, created_at, password_hash
	           FROM app.users WHERE email=$1`
	var u domain.User
	var hash string
	if err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	return &u, hash, nil
}
