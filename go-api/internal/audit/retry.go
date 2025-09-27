package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type DBExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type RetryConfig struct {
	Brokers      []string
	DLQTopic     string
	Group        string
	MaxAttempts  int
	BackoffStart time.Duration
	Logger       *slog.Logger
	DB           DBExec
}

const hdrAttempts = "x-attempts"

func RunRetry(ctx context.Context, cfg RetryConfig) error {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.Group,
		Topic:          cfg.DLQTopic,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	defer r.Close()

	w := &kafka.Writer{Addr: kafka.TCP(cfg.Brokers...), Topic: cfg.DLQTopic, RequiredAcks: kafka.RequireAll}
	defer w.Close()

	backoff0 := cfg.BackoffStart
	if backoff0 <= 0 {
		backoff0 = 2 * time.Second
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Error("dlq_read", "err", err)
			continue
		}

		attempt := 0
		for _, h := range m.Headers {
			if h.Key == hdrAttempts {
				if v, e := strconv.Atoi(string(h.Value)); e == nil {
					attempt = v
				}
				break
			}
		}

		if err := insertAudit(ctx, cfg.DB, m.Value); err == nil {
			continue
		}

		if attempt+1 >= cfg.MaxAttempts {
			if err := parkDLQ(ctx, cfg.DB, m.Value, attempt+1); err != nil {
				log.Error("dlq_park_insert", "err", err)
			}
			continue
		}

		time.Sleep(backoff0 * time.Duration(attempt+1))
		m.Headers = upsertHeader(m.Headers, hdrAttempts, []byte(strconv.Itoa(attempt+1)))
		if err := w.WriteMessages(ctx, kafka.Message{Key: m.Key, Value: m.Value, Headers: m.Headers}); err != nil {
			log.Error("dlq_reenqueue", "err", err)
		}
	}
}

func upsertHeader(hs []kafka.Header, k string, v []byte) []kafka.Header {
	for i := range hs {
		if hs[i].Key == k {
			hs[i].Value = v
			return hs
		}
	}
	return append(hs, kafka.Header{Key: k, Value: v})
}

func insertAudit(ctx context.Context, db DBExec, payload []byte) error {
	var ev itemEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO app.item_audit(evt_type, payload) VALUES ($1,$2)`,
		ev.Type, json.RawMessage(payload),
	)
	return err
}

func parkDLQ(ctx context.Context, db DBExec, payload []byte, attempts int) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO app.item_audit_parking(payload, attempts) VALUES ($1,$2)`,
		json.RawMessage(payload), attempts,
	)
	return err
}
