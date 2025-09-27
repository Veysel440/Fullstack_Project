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

func Test_List_Integration(t *testing.T) {
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
	INSERT INTO app.items(name,price) VALUES('A',1.2),('B',3.4);
	`)

	r := repo.NewItemRepo(db)
	list, total, err := r.ListPagedSortedWithTotal(ctx, 10, 0, "id,desc", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("unexpected: total=%d len=%d", total, len(list))
	}
}
