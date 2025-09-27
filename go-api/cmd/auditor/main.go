package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"fullstack-oracle/go-api/internal/audit"
	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
	"fullstack-oracle/go-api/internal/migrate"
)

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.FromEnv()
	d, err := db.Open(cfg)
	if err != nil {
		logger.Error("db_open", "err", err)
		os.Exit(1)
	}
	if err := migrate.Up(context.Background(), d); err != nil {
		logger.Error("migrate_up", "err", err)
		os.Exit(1)
	}

	cons, err := audit.NewConsumer(audit.Config{
		Brokers:   splitCSV(os.Getenv("KAFKA_BROKERS")),
		Topic:     envOr("KAFKA_ITEMS_TOPIC", "item"),
		Group:     envOr("KAFKA_GROUP", "items-auditor"),
		DeadTopic: envOr("KAFKA_DLQ_TOPIC", "item-dlq"),
		DB:        d,
		Logger:    logger,
	})
	if err != nil {
		logger.Error("consumer_new", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := cons.Run(ctx); err != nil {
		logger.Error("consumer_run", "err", err)
		os.Exit(1)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
