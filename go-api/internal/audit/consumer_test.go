package audit

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandle_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	c := &Consumer{db: db}

	payload := []byte(`{"type":"item.created","item":{"id":1,"name":"X"}}`)

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO app.item_audit(evt_type, payload) VALUES ($1,$2)`)).
		WithArgs("item.created", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := c.handle(context.Background(), payload); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandle_BadJSON(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	c := &Consumer{db: db}

	if err := c.handle(context.Background(), []byte(`{oops}`)); err == nil {
		t.Fatal("want error, got nil")
	}
}
