package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fullstack-oracle/go-api/internal/audit"
	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
	"fullstack-oracle/go-api/internal/migrate"
)

func main() {
	cfg := config.FromEnv()

	d, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrate.Up(context.Background(), d); err != nil {
		log.Fatal(err)
	}

	cons, err := audit.NewConsumer(audit.Config{
		Brokers:   []string{os.Getenv("KAFKA_BROKERS")},
		Topic:     "item",
		Group:     "items-auditor",
		DeadTopic: "item-dlq",
		DB:        d,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := cons.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
