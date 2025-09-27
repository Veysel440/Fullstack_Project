package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fullstack-oracle/go-api/internal/audit"
	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
)

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	ss := strings.Split(s, ",")
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	return ss
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.FromEnv()

	d, err := db.Open(cfg)
	if err != nil {
		logger.Error("db_open", "err", err)
		os.Exit(1)
	}

	rc := audit.RetryConfig{
		Brokers:      splitCSV(os.Getenv("KAFKA_BROKERS")),
		DLQTopic:     envOr("KAFKA_DLQ_TOPIC", "item-dlq"),
		Group:        envOr("KAFKA_RETRY_GROUP", "items-auditor-retry"),
		MaxAttempts:  5,
		BackoffStart: 2 * time.Second,
		Logger:       logger,
		DB:           d,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := audit.RunRetry(ctx, rc); err != nil {
		logger.Error("retry_run", "err", err)
		os.Exit(1)
	}
}
