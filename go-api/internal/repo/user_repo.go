package repo

import (
	"context"
	"database/sql"
	"errors"

	"fullstack-oracle/go-api/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrConflict = errors.New("conflict")

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	const q = `SELECT id, email, role, created_at, password_hash FROM app.users WHERE email=$1`
	var u domain.User
	var hash string
	if err := r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	return &u, hash, nil
}

func (r *UserRepo) Create(ctx context.Context, email, hash, role string) (*domain.User, error) {
	const q = `INSERT INTO app.users(email,password_hash,role) VALUES($1,$2,$3)
	           RETURNING id,email,role,created_at`
	var u domain.User
	if err := r.db.QueryRowContext(ctx, q, email, hash, role).
		Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, err
	}
	return &u, nil
}
