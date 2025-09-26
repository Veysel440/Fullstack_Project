package repo_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
)

func TestCreateList_Integration(t *testing.T) {
	ctx := context.Background()
	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pg.Terminate(ctx) }()

	uri, _ := pg.ConnectionString(ctx, "sslmode=disable")
	db, err := sql.Open("pgx", uri)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, _ = db.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS app;
		CREATE TABLE IF NOT EXISTS app.items(
		  id bigserial PRIMARY KEY,
		  name varchar(100) NOT NULL,
		  price numeric(10,2) NOT NULL DEFAULT 0,
		  created_at timestamptz NOT NULL DEFAULT now()
		);`)

	r := repo.NewItemRepo(db)
	if _, err = r.Create(ctx, domain.CreateItemDTO{Name: "X", Price: 1.23}); err != nil {
		t.Fatal(err)
	}

	list, err := r.List(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("list err=%v len=%d", err, len(list))
	}
}
