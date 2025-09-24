package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var fs embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	c, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := goose.UpContext(c, db, "sql"); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
