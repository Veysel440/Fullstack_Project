package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
	hh "fullstack-oracle/go-api/internal/http"
	"fullstack-oracle/go-api/internal/migrate"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"
)

func main() {
	cfg := config.FromEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if cfg.SentryDSN != "" {
		_ = sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDSN,
			Environment: cfg.SentryEnv,
		})
		defer sentry.Flush(2 * time.Second)
	}

	d, err := db.Open(cfg)
	if err != nil {
		logger.Error("db_open", "err", err)
		os.Exit(1)
	}

	if err := migrate.Up(context.Background(), d); err != nil {
		logger.Error("migrate_up", "err", err)
		os.Exit(1)
	}

	rp := repo.NewItemRepo(d)
	sv := service.NewItemService(rp)
	h := &hh.Handlers{S: sv}

	rl := hh.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	app := hh.Router(h, hh.CORS(cfg.CORSOrigins), logger, rl)

	addr := ":" + cfg.Port
	logger.Info("api_listen", "addr", addr)
	if err := http.ListenAndServe(addr, app); err != nil {
		logger.Error("api_exit", "err", err)
	}
}
