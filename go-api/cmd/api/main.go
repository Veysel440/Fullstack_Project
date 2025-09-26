package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"

	"fullstack-oracle/go-api/internal/cache"
	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
	"fullstack-oracle/go-api/internal/events"
	hh "fullstack-oracle/go-api/internal/http"
	"fullstack-oracle/go-api/internal/migrate"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"
)

func main() {
	cfg := config.FromEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if cfg.SentryDSN != "" {
		_ = sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN, Environment: cfg.SentryEnv})
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

	c := cache.New()
	if c != nil {
		defer c.Close()
	}
	ev := events.NewWriter()
	if ev != nil {
		defer ev.Close()
	}

	itemRepo := repo.NewItemRepo(d)
	itemSvc := service.NewItemService(itemRepo, nil)
	h := &hh.Handlers{S: itemSvc}

	rl := hh.NewRateLimiter(float64(cfg.RateLimitRPS), cfg.RateLimitBurst)

	var jwtv *hh.JWTVerifier
	var ah *hh.AuthHandlers
	if cfg.JWTAccessSecret != "" && cfg.JWTRefreshSecret != "" {
		userRepo := repo.NewUserRepo(d)
		authSvc := service.NewAuthService(cfg, userRepo, c)
		ah = &hh.AuthHandlers{Cfg: cfg, S: authSvc, Val: validator.New()}
		jwtv = &hh.JWTVerifier{
			AccessSecret:  []byte(cfg.JWTAccessSecret),
			RefreshSecret: []byte(cfg.JWTRefreshSecret),
			Logger:        logger,
		}
	}

	corsMW := hh.CORS(strings.Join(cfg.CORSOrigins, ","))
	app := hh.Router(h, corsMW, logger, rl, jwtv, ah)

	addr := ":" + cfg.Port
	logger.Info("api_listen", "addr", addr)
	if err := http.ListenAndServe(addr, app); err != nil {
		logger.Error("api_exit", "err", err)
	}
}
