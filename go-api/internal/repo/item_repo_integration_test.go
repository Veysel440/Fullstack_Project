//go:build integration

package repo_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
)

func TestCreateList_Integration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("run pg container: %v", err)
	}
	defer func() { _ = pg.Terminate(context.Background()) }()

	uri, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}

	db, err := sql.Open("pgx", uri)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// minimal şema
	_, _ = db.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS app;
		CREATE TABLE IF NOT EXISTS app.items(
		  id bigserial PRIMARY KEY,
		  name varchar(100) NOT NULL,
		  price numeric(10,2) NOT NULL DEFAULT 0,
		  created_at timestamptz NOT NULL DEFAULT now()
		);
	`)

	r := repo.NewItemRepo(db)

	if _, err := r.Create(ctx, domain.CreateItemDTO{Name: "X", Price: 1.23}); err != nil {
		t.Fatalf("create: %v", err)
	}
	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least 1 row")
	}
}
