package repo_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

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
	defer pg.Terminate(ctx)

	uri, _ := pg.ConnectionString(ctx, "sslmode=disable")
	db, err := sql.Open("pgx", uri)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, _ = db.ExecContext(ctx, `
		CREATE SCHEMA app;
		CREATE TABLE app.items(
		  id bigserial PRIMARY KEY,
		  name varchar(100) NOT NULL,
		  price numeric(10,2) NOT NULL DEFAULT 0,
		  created_at timestamptz NOT NULL DEFAULT now()
		);
	`)

	r := repo.NewItemRepo(db)
	_, err = r.Create(ctx, repo.CreateItemDTO{Name: "X", Price: 1.23})
	if err != nil {
		t.Fatal(err)
	}

	list, err := r.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal("no data")
	}

	_ = time.Now()
}
