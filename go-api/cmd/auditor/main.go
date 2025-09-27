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
	ss := strings.Split(s, ",")
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	return ss
}

func main() {
	log := slog.Default()
	cfg := config.FromEnv()

	d, err := db.Open(cfg)
	if err != nil {
		log.Error("db_open", "err", err)
		os.Exit(1)
	}
	if err := migrate.Up(context.Background(), d); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}

	cons, err := audit.NewConsumer(audit.Config{
		Brokers:   splitCSV(os.Getenv("KAFKA_BROKERS")),
		Topic:     getenv("KAFKA_ITEMS_TOPIC", "item"),
		Group:     getenv("KAFKA_GROUP", "items-auditor"),
		DeadTopic: getenv("KAFKA_DLQ_TOPIC", "item-dlq"),
		DB:        d,
		Logger:    log,
	})
	if err != nil {
		log.Error("consumer_init", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := cons.Run(ctx); err != nil {
		log.Error("run", "err", err)
		os.Exit(1)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
