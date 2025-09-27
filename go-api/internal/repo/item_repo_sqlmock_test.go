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

func Test_ListPagedSortedWithTotal(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := repo.NewItemRepo(db)
	
	rows := sqlmock.NewRows([]string{"id", "name", "price", "created_at"}).
		AddRow(int64(2), "B", 20.0, time.Now()).
		AddRow(int64(1), "A", 10.0, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`FROM app.items`)).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM app.items`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	_, total, err := r.ListPagedSortedWithTotal(context.Background(), 10, 0, "id,desc", "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if total != 2 {
		t.Fatalf("want total=2 got %d", total)
	}
}
