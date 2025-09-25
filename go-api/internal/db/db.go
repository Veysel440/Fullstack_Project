package db

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"fullstack-oracle/go-api/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func getenvInt(k string, def int) int {
	if s := os.Getenv(k); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return def
}
func getenvDurMin(k string, defMin int) time.Duration {
	if s := os.Getenv(k); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return time.Duration(v) * time.Minute
		}
	}
	return time.Duration(defMin) * time.Minute
}

func Open(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(getenvInt("DB_MAX_OPEN", 10))
	db.SetMaxIdleConns(getenvInt("DB_MAX_IDLE", 5))
	db.SetConnMaxLifetime(getenvDurMin("DB_CONN_MAX_LIFETIME_MIN", 30))
	return db, db.Ping()
}
