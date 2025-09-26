package repo_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"fullstack-oracle/go-api/internal/repo"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGet_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	r := repo.NewItemRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "price", "created_at"}).
		AddRow(int64(1), "A", 10.0, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,name,price,created_at FROM app.items WHERE id=$1`)).
		WithArgs(int64(1)).WillReturnRows(rows)

	if _, err := r.Get(context.Background(), 1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
